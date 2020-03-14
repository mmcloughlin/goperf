package fs

import (
	"bytes"
	"context"
	"testing"

	"github.com/mmcloughlin/cb/internal/test"
)

func TestWriteReadFile(t *testing.T) {
	d, clean := test.TempDir(t)
	defer clean()

	filesystems := map[string]Interface{
		"mem":   NewMem(),
		"local": NewLocal(d),
	}
	for name, fs := range filesystems {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			expect := []byte("Hello, World!\n")
			if err := WriteFile(ctx, fs, "greeting.txt", expect); err != nil {
				t.Fatal(err)
			}
			got, err := ReadFile(ctx, fs, "greeting.txt")
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(expect, got) {
				t.Fatal("mismatch")
			}
		})
	}
}
