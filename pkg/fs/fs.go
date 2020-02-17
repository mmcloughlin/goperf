package fs

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

// Interface is a filesystem abstraction.
type Interface interface {
	Open(ctx context.Context, name string) (io.WriteCloser, error)
}

type local struct {
	root string
}

// NewLocal returns a filesystem rooted at the provided path on the local
// system.
func NewLocal(root string) Interface {
	return &local{root: root}
}

func (l *local) Open(ctx context.Context, name string) (io.WriteCloser, error) {
	path := filepath.Join(l.root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return nil, err
	}
	return os.Create(path)
}

// Discard stores nothing.
var Discard Interface = discard{}

type discard struct{}

func (discard) Open(context.Context, string) (io.WriteCloser, error) {
	return devnull{}, nil
}

type devnull struct{}

func (devnull) Write(p []byte) (int, error) { return len(p), nil }
func (devnull) Close() error                { return nil }
