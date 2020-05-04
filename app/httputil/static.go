package httputil

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/mmcloughlin/cb/pkg/fs"
)

// Static serves static content from a filesystem. Filenames are extended with a
// content hash for cache busting.
type Static struct {
	fs  fs.Readable
	cc  CacheControl
	log *zap.Logger

	mu sync.Mutex
	v  map[string]string
}

// NewStatic builds a static file HTTP handler.
func NewStatic(filesys fs.Readable) *Static {
	return &Static{
		fs:  filesys,
		cc:  CacheControlImmutable,
		log: zap.NewNop(),
		v:   map[string]string{},
	}
}

// SetCacheControl configures cache control headers.
func (s *Static) SetCacheControl(cc CacheControl) { s.cc = cc }

// SetLogger configures logging output from the static handler.
func (s *Static) SetLogger(l *zap.Logger) { s.log = l.Named("static") }

// Path returns the versioned path for name, intended for cache busting.
func (s *Static) Path(ctx context.Context, name string) (string, error) {
	v, err := s.version(ctx, name)
	if err != nil {
		return "", err
	}
	return versionedPath(name, v), nil
}

// HandleRequest serves a request.
func (s *Static) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	name, v := parseNameVersion(r.URL.Path)
	if v == "" {
		s.log.Info("unversioned request")
		return NotFound()
	}

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
	if v != expect {
		s.log.Info("version mismatch", zap.String("v", v), zap.String("expect", expect))
		return NotFound()
	}

	// Cache control.
	w.Header().Set("Cache-Control", s.cc.String())

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
