package fs

import (
	"bytes"
	"context"
	"testing"
)

func TestWriteReadFile(t *testing.T) {
	FilesystemTest(t, func(ctx context.Context, t *testing.T, fs Interface) {
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
