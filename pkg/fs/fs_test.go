package fs

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

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
	if err != os.ErrNotExist {
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
