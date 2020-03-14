package fs

import (
	"context"
	"io/ioutil"
)

// ReadFile reads the named file from the given filesystem and returns the contents.
func ReadFile(ctx context.Context, fs Readable, name string) ([]byte, error) {
	r, err := fs.Open(ctx, name)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if err := r.Close(); err != nil {
		return nil, err
	}

	return data, nil
}

// WriteFile writes data to the named file in the supplied filesystem.
func WriteFile(ctx context.Context, fs Writable, name string, data []byte) error {
	w, err := fs.Create(ctx, name)
	if err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}

	return w.Close()
}
