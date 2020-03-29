package fs

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestSub(t *testing.T) {
	tree := map[string]string{
		"a":          "a",
		"d0/a":       "a",
		"d0/d1/a":    "a",
		"d0/d1/d2/a": "a",
	}
	FilesystemTest(t, func(ctx context.Context, t *testing.T, fs Interface) {
		// Populate.
		for path, content := range tree {
			if err := WriteFile(ctx, fs, path, []byte(content)); err != nil {
				t.Fatal(err)
			}
		}

		// Create sub-filesystem.
		sub := NewSub(fs, "d0")

		// Ensure everything in the listing can be opened.
		files, err := sub.List(ctx, "")
		if err != nil {
			t.Fatal(err)
		}

		for _, file := range files {
			_, err := ReadFile(ctx, sub, file.Path)
			if err != nil {
				t.Fatal(err)
			}
		}
	})
}

func TestMem(t *testing.T) {
	ctx := context.Background()
	m := NewMem()

	// Create a new file.
	w, err := m.Create(ctx, "greeting.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err = fmt.Fprintln(w, "Hello, World!"); err != nil {
		t.Fatal(err)
	}

	// File should on exist "on disk" until close.
	_, err = m.Open(ctx, "greeting.txt")
	if err != ErrNotExist {
		t.Fatal("expected file to not exist")
	}

	// Close it to flush.
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	// Double close is an error.
	if err := w.Close(); err == nil {
		t.Fatal("expected error on double close")
	}

	// Read it back.
	r, err := m.Open(ctx, "greeting.txt")
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "Hello, World!\n" {
		t.Fatalf(`incorrect file contents: %q`, string(b))
	}

	// Close it.
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMemMultiRead(t *testing.T) {
	ctx := context.Background()
	expect := []byte("Hello, World!\n")
	m := NewMem()

	if err := WriteFile(ctx, m, "greeting.txt", expect); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		got, err := ReadFile(ctx, m, "greeting.txt")
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(expect, got) {
			t.Fatalf("read #%d: mismatch", i+1)
		}
	}
}
