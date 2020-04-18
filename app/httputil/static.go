package httputil

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/mmcloughlin/cb/pkg/fs"
)

// CacheControl specifies Cache-Control header options.
type CacheControl struct {
	MaxAge       time.Duration
	SharedMaxAge time.Duration
	Directives   []string
}

func (c CacheControl) String() string {
	directives := c.Directives
	if c.MaxAge != 0 {
		directives = append(directives, fmt.Sprintf("max-age=%d", c.MaxAge/time.Second))
	}
	if c.SharedMaxAge != 0 {
		directives = append(directives, fmt.Sprintf("s-maxage=%d", c.SharedMaxAge/time.Second))
	}
	return strings.Join(directives, ", ")
}

// Static serves static content froma filesystem.
type Static struct {
	fs  fs.Readable
	log *zap.Logger

	defaultCache   CacheControl
	versionedCache CacheControl

	mu sync.Mutex
	v  map[string]string
}

func NewStatic(filesys fs.Readable) *Static {
	return &Static{
		fs:  filesys,
		log: zap.NewNop(),
		defaultCache: CacheControl{
			MaxAge:     24 * time.Hour,
			Directives: []string{"public"},
		},
		versionedCache: CacheControl{
			MaxAge:       24 * time.Hour,
			SharedMaxAge: 365 * 24 * time.Hour,
			Directives:   []string{"public", "immutable"},
		},
		v: map[string]string{},
	}
}

func (s *Static) SetLogger(l *zap.Logger) { s.log = l.Named("static") }

func (s *Static) SetDefaultCacheControl(cc CacheControl) {
	s.defaultCache = cc
}

func (s *Static) SetVersionedCacheControl(cc CacheControl) {
	s.versionedCache = cc
}

// Path returns the versioned path for name, intended for cache busting.
func (s *Static) Path(ctx context.Context, name string) (string, error) {
	v, err := s.version(ctx, name)
	if err != nil {
		return "", err
	}
	return versionedPath(name, v), nil
}

func (s *Static) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	name, v := parseNameVersion(r.URL.Path)

	// Fetch the file.
	info, err := s.fs.Stat(ctx, name)
	if err != nil {
		s.log.Info("stat file error", zap.Error(err))
		return NotFound()
	}

	b, expect, err := s.read(ctx, name)
	if err != nil {
		s.log.Info("file read error", zap.Error(err))
		return NotFound()
	}
	if v != "" && v != expect {
		s.log.Info("version mismatch", zap.String("v", v), zap.String("expect", expect))
		return NotFound()
	}

	// Cache control.
	cc := s.defaultCache
	if v != "" {
		cc = s.versionedCache
	}
	w.Header().Set("Cache-Control", cc.String())

	http.ServeContent(w, r, name, info.ModTime, bytes.NewReader(b))
	return nil
}

// read the file and save the version if not already known.
func (s *Static) read(ctx context.Context, name string) ([]byte, string, error) {
	// Read from filesystem.
	b, err := fs.ReadFile(ctx, s.fs, name)
	if err != nil {
		return nil, "", err
	}

	// Cache version.
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.v[name]; !ok {
		s.v[name] = version(b)
	}

	return b, s.v[name], nil
}

// version returns the version code for the named file.
func (s *Static) version(ctx context.Context, name string) (string, error) {
	// Check to see if we already have it cached.
	s.mu.Lock()
	v, ok := s.v[name]
	s.mu.Unlock()

	if ok {
		return v, nil
	}

	// Compute and store it by reading the file.
	_, v, err := s.read(ctx, name)
	if err != nil {
		return "", err
	}

	return v, nil
}

const versionLen = 12

func version(b []byte) string {
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])[:versionLen]
}

func versionedPath(name, version string) string {
	ext := filepath.Ext(name)
	return name[:len(name)-len(ext)] + "." + version[:versionLen] + ext
}

func parseNameVersion(path string) (string, string) {
	// Find the last two ".".
	e := strings.LastIndexByte(path, '.')
	if e < 0 {
		return path, ""
	}
	h := strings.LastIndexByte(path[:e], '.')
	if h < 0 {
		return path, ""
	}

	// Extract hash.
	if e-h != versionLen+1 {
		return path, ""
	}

	return path[:h] + path[e:], path[h+1 : e]
}
