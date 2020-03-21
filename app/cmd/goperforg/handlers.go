package main

import (
	"fmt"
	"net/http"

	"github.com/mmcloughlin/cb/app/service"
)

type Handlers struct {
	srv service.Service
	mux *http.ServeMux
}

func NewHandlers(srv service.Service) *Handlers {
	h := &Handlers{
		srv: srv,
		mux: http.NewServeMux(),
	}

	h.mux.HandleFunc("/", h.Index)

	return h
}

func (h *Handlers) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Fetch modules.
	mods, err := h.srv.ListModules(ctx)
	if err != nil {
		httperror(w, err)
		return
	}

	// Write response.
	for _, mod := range mods {
		_, err := fmt.Fprintln(w, mod.UUID(), mod.Path, mod.Version)
		if err != nil {
			httperror(w, err)
			return
		}
	}
}

func (h *Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func httperror(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
