package main

import (
	"bytes"
	"net/http"

	"github.com/mmcloughlin/cb/pkg/fs"
)

func NewStatic(filesys fs.Readable) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		name := r.URL.Path

		info, err := filesys.Stat(ctx, name)
		if err != nil {
			Error(w, err)
			return
		}

		b, err := fs.ReadFile(ctx, filesys, name)
		if err != nil {
			Error(w, err)
			return
		}

		http.ServeContent(w, r, name, info.ModTime, bytes.NewReader(b))
	})
}

func Error(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
