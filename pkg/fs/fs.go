package fs

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

// FileInfo describes a file.
type FileInfo struct {
	Path string // path to the file relative to the filesystem
	Size int64  // size in bytes
}

// Writable can create named files.
type Writable interface {
	Create(ctx context.Context, name string) (io.WriteCloser, error)
	Remove(ctx context.Context, name string) error
}

// Readable can read from named files.
type Readable interface {
	Open(ctx context.Context, name string) (io.ReadCloser, error)
	List(ctx context.Context) ([]*FileInfo, error)
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

func (l *local) List(ctx context.Context) ([]*FileInfo, error) {
	var files []*FileInfo
	err := filepath.Walk(l.root, func(path string, info os.FileInfo, err error) error {
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
			Path: rel,
			Size: info.Size(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// Discard stores nothing.
var Discard Writable = discard{}

type discard struct{}

func (discard) Create(context.Context, string) (io.WriteCloser, error) {
	return devnull{}, nil
}

func (discard) Remove(ctx context.Context, name string) error {
	return nil
}

type devnull struct{}

func (devnull) Write(p []byte) (int, error) { return len(p), nil }
func (devnull) Close() error                { return nil }

type mem struct {
	files map[string][]byte
	mu    sync.RWMutex
}

// NewMem builds an in-memory filesystem.
func NewMem() Interface {
	return &mem{
		files: map[string][]byte{},
	}
}

func (m *mem) Create(_ context.Context, name string) (io.WriteCloser, error) {
	return &memfile{
		Buffer: bytes.NewBuffer(nil),
		name:   name,
		fs:     m,
	}, nil
}

func (m *mem) Remove(_ context.Context, name string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, ok := m.files[name]; !ok {
		return os.ErrNotExist
	}

	delete(m.files, name)
	return nil
}

func (m *mem) Open(_ context.Context, name string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	b, ok := m.files[name]
	if !ok {
		return nil, os.ErrNotExist
	}

	return ioutil.NopCloser(bytes.NewBuffer(b)), nil
}

func (m *mem) List(_ context.Context) ([]*FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	files := make([]*FileInfo, 0, len(m.files))
	for path, data := range m.files {
		files = append(files, &FileInfo{
			Path: path,
			Size: int64(len(data)),
		})
	}

	return files, nil
}

type memfile struct {
	*bytes.Buffer
	name string
	fs   *mem
}

func (f *memfile) Close() error {
	if f.fs == nil {
		return errors.New("already closed")
	}
	f.fs.mu.Lock()
	defer f.fs.mu.Unlock()
	f.fs.files[f.name] = f.Buffer.Bytes()
	f.fs = nil
	return nil
}
