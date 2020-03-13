package fs

import (
	"context"
	"io"
	"os"
	"path/filepath"
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
