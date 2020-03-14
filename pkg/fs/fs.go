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

// Writable can create named files.
type Writable interface {
	Create(ctx context.Context, name string) (io.WriteCloser, error)
}

// Readable can read from named files.
type Readable interface {
	Open(ctx context.Context, name string) (io.ReadCloser, error)
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

func (l *local) Open(ctx context.Context, name string) (io.ReadCloser, error) {
	path := filepath.Join(l.root, name)
	return os.Open(path)
}

// Discard stores nothing.
var Discard Writable = discard{}

type discard struct{}

func (discard) Create(context.Context, string) (io.WriteCloser, error) {
	return devnull{}, nil
}

type devnull struct{}

func (devnull) Write(p []byte) (int, error) { return len(p), nil }
func (devnull) Close() error                { return nil }

type mem struct {
	files map[string]io.ReadCloser
	mu    sync.RWMutex
}

// NewMem builds an in-memory filesystem.
func NewMem() Interface {
	return &mem{
		files: map[string]io.ReadCloser{},
	}
}

func (m *mem) Create(_ context.Context, name string) (io.WriteCloser, error) {
	return &memfile{
		Buffer: bytes.NewBuffer(nil),
		name:   name,
		fs:     m,
	}, nil
}

func (m *mem) Open(_ context.Context, name string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	f, ok := m.files[name]
	if !ok {
		return nil, os.ErrNotExist
	}

	return f, nil
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
	f.fs.files[f.name] = ioutil.NopCloser(f.Buffer)
	f.fs = nil
	return nil
}
