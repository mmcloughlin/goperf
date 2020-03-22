package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/app/service"
	"github.com/mmcloughlin/cb/pkg/fs"
)

type Handlers struct {
	srv    service.Service
	tmplfs fs.Readable

	mux *http.ServeMux
}

type Option func(*Handlers)

func WithTemplateFileSystem(r fs.Readable) Option {
	return func(h *Handlers) { h.tmplfs = r }
}

func NewHandlers(srv service.Service, opts ...Option) *Handlers {
	h := &Handlers{
		srv:    srv,
		tmplfs: AssetFileSystem(),
		mux:    http.NewServeMux(),
	}
	for _, opt := range opts {
		opt(h)
	}

	h.mux.HandleFunc("/mods/", h.Modules)
	h.mux.HandleFunc("/mod/", h.Module)

	return h
}

func (h *Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handlers) Modules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Fetch modules.
	mods, err := h.srv.ListModules(ctx)
	if err != nil {
		httperror(w, err)
		return
	}

	// Write response.
	h.render(ctx, w, "mods.html", map[string]interface{}{
		"Modules": mods,
	})
}

func (h *Handlers) Module(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse UUID.
	id, err := parseuuid(r.URL.Path, "/mod/")
	if err != nil {
		httperror(w, err)
		return
	}

	// Fetch module.
	mod, err := h.srv.FindModuleByUUID(ctx, id)
	if err != nil {
		httperror(w, err)
		return
	}

	// Write response.
	h.render(ctx, w, "mod.html", mod)
}

func (h *Handlers) render(ctx context.Context, w http.ResponseWriter, name string, data interface{}) {
	tmpl, err := fs.ReadFile(ctx, h.tmplfs, name)
	if err != nil {
		httperror(w, err)
		return
	}

	t, err := template.New(name).Parse(string(tmpl))
	if err != nil {
		httperror(w, err)
		return
	}

	if err := t.Execute(w, data); err != nil {
		if err != nil {
			httperror(w, err)
			return
		}
	}
}

func httperror(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func parseuuid(path, prefix string) (uuid.UUID, error) {
	if !strings.HasPrefix(path, prefix) {
		return uuid.Nil, fmt.Errorf("path %q expected to have prefix %q", path, prefix)
	}
	return uuid.Parse(path[len(prefix):])
}
