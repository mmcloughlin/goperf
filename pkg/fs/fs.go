package fs

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ErrNotExist is returned when a file does not exist.
var ErrNotExist = errors.New("file does not exist")

// FileInfo describes a file.
type FileInfo struct {
	Path    string    // path to the file relative to the filesystem
	Size    int64     // size in bytes
	ModTime time.Time // modification time
}

// Writable can create named files.
type Writable interface {
	Create(ctx context.Context, name string) (io.WriteCloser, error)
	Remove(ctx context.Context, name string) error
}

// Readable can read from named files.
type Readable interface {
	Open(ctx context.Context, name string) (io.ReadCloser, error)
	List(ctx context.Context, prefix string) ([]*FileInfo, error)
}

// Interface is a filesystem abstraction.
type Interface interface {
	Writable
	Readable
}

type local struct {
	root string
}

// NewLocal returns a filesystem rooted at the provided path on the local
// system.
func NewLocal(root string) Interface {
	return &local{root: root}
}

func (l *local) Create(ctx context.Context, name string) (io.WriteCloser, error) {
	path := filepath.Join(l.root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return nil, err
	}
	return os.Create(path)
}

func (l *local) Remove(ctx context.Context, name string) error {
	path := filepath.Join(l.root, name)
	return os.Remove(path)
}

func (l *local) Open(ctx context.Context, name string) (io.ReadCloser, error) {
	path := filepath.Join(l.root, name)
	return os.Open(path)
}

func (l *local) List(ctx context.Context, prefix string) ([]*FileInfo, error) {
	var files []*FileInfo
	err := filepath.Walk(filepath.Join(l.root, prefix), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(l.root, path)
		if err != nil {
			return err
		}
		files = append(files, &FileInfo{
			Path:    rel,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// Null contains no files and discards writes.
var Null Interface = null{}

type null struct{}

func (null) Create(context.Context, string) (io.WriteCloser, error) {
	return discard{}, nil
}

func (null) Remove(ctx context.Context, name string) error {
	return nil
}

func (null) Open(ctx context.Context, name string) (io.ReadCloser, error) {
	return nil, ErrNotExist
}

func (null) List(ctx context.Context, prefix string) ([]*FileInfo, error) {
	return []*FileInfo{}, nil
}

type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }
func (discard) Close() error                { return nil }

type sub struct {
	fs     Interface
	prefix string
}

// NewSub returns the sub-filesystem of fs rooted at prefix.
func NewSub(fs Interface, prefix string) Interface {
	return &sub{
		fs:     fs,
		prefix: prefix,
	}
}

func (s *sub) Create(ctx context.Context, name string) (io.WriteCloser, error) {
	return s.fs.Create(ctx, s.path(name))
}

func (s *sub) Remove(ctx context.Context, name string) error {
	return s.fs.Remove(ctx, s.path(name))
}

func (s *sub) Open(ctx context.Context, name string) (io.ReadCloser, error) {
	return s.fs.Open(ctx, s.path(name))
}

func (s *sub) List(ctx context.Context, prefix string) ([]*FileInfo, error) {
	return s.fs.List(ctx, s.path(prefix))
}

func (s *sub) path(name string) string {
	return filepath.Join(s.prefix, name)
}

type mem struct {
	files map[string]memfile
	mu    sync.RWMutex
}

type memfile struct {
	data    []byte
	modtime time.Time
}

// NewMem builds an in-memory filesystem.
func NewMem() Interface {
	return NewMemWithFiles(map[string][]byte{})
}

// NewMemWithFiles creates an in-memory filesystem initialized with the given files.
func NewMemWithFiles(files map[string][]byte) Interface {
	m := &mem{
		files: map[string]memfile{},
	}
	for name, data := range files {
		m.files[name] = memfile{
			data:    data,
			modtime: time.Now(),
		}
	}
	return m
}

func (m *mem) Create(_ context.Context, name string) (io.WriteCloser, error) {
	return &memwriter{
		Buffer: bytes.NewBuffer(nil),
		name:   name,
		fs:     m,
	}, nil
}

func (m *mem) Remove(_ context.Context, name string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, ok := m.files[name]; !ok {
		return ErrNotExist
	}

	delete(m.files, name)
	return nil
}

func (m *mem) Open(_ context.Context, name string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	f, ok := m.files[name]
	if !ok {
		return nil, ErrNotExist
	}

	return ioutil.NopCloser(bytes.NewBuffer(f.data)), nil
}

func (m *mem) List(_ context.Context, prefix string) ([]*FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	files := make([]*FileInfo, 0, len(m.files))
	for path, file := range m.files {
		if !strings.HasPrefix(path, prefix) {
			continue
		}
		files = append(files, &FileInfo{
			Path:    path,
			Size:    int64(len(file.data)),
			ModTime: file.modtime,
		})
	}

	return files, nil
}

type memwriter struct {
	*bytes.Buffer
	name string
	fs   *mem
}

func (w *memwriter) Close() error {
	if w.fs == nil {
		return errors.New("already closed")
	}
	w.fs.mu.Lock()
	defer w.fs.mu.Unlock()
	w.fs.files[w.name] = memfile{
		data:    w.Bytes(),
		modtime: time.Now(),
	}
	w.fs = nil
	return nil
}
