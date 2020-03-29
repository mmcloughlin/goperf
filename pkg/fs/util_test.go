package fs

import (
	"context"
	"testing"

	"github.com/mmcloughlin/cb/internal/test"
)

func FilesystemTest(t *testing.T, f func(ctx context.Context, t *testing.T, fs Interface)) {
	d, clean := test.TempDir(t)
	defer clean()

	filesystems := map[string]Interface{
		"mem":   NewMem(),
		"local": NewLocal(d),
	}

	for name, fs := range filesystems {
		fs := fs // scopelint
		t.Run(name, func(t *testing.T) {
			f(context.Background(), t, fs)
		})
	}
}
