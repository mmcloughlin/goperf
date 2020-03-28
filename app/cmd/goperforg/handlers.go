package main

import (
	"bufio"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/app/db"
	"github.com/mmcloughlin/cb/pkg/fs"
)

type Handlers struct {
	srv    *db.DB
	tmplfs fs.Readable
	static fs.Readable
	datafs fs.Readable

	mux       *http.ServeMux
	templates *template.Template
}

type Option func(*Handlers)

func WithTemplateFileSystem(r fs.Readable) Option {
	return func(h *Handlers) { h.tmplfs = r }
}

func WithStaticFileSystem(r fs.Readable) Option {
	return func(h *Handlers) { h.static = r }
}

func WithDataFileSystem(r fs.Readable) Option {
	return func(h *Handlers) { h.datafs = r }
}

func NewHandlers(d *db.DB, opts ...Option) *Handlers {
	// Configure.
	h := &Handlers{
		srv:       d,
		tmplfs:    TemplateFileSystem,
		static:    StaticFileSystem,
		datafs:    fs.Null,
		mux:       http.NewServeMux(),
		templates: template.New(""),
	}
	for _, opt := range opts {
		opt(h)
	}

	// Setup mux.
	h.mux.HandleFunc("/mods/", h.Modules)
	h.mux.HandleFunc("/mod/", h.Module)
	h.mux.HandleFunc("/pkg/", h.Package)
	h.mux.HandleFunc("/bench/", h.Benchmark)
	h.mux.HandleFunc("/file/", h.File)

	static := NewStatic(h.static)
	h.mux.Handle("/static/", http.StripPrefix("/static/", static))

	return h
}

func (h *Handlers) Init(ctx context.Context) error {
	files, err := h.tmplfs.List(ctx, "")
	if err != nil {
		return err
	}

	for _, file := range files {
		b, err := fs.ReadFile(ctx, h.tmplfs, file.Path)
		if err != nil {
			return err
		}

		if _, err := h.templates.Parse(string(b)); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handlers) Modules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Fetch modules.
	mods, err := h.srv.ListModules(ctx)
	if err != nil {
		Error(w, err)
		return
	}

	// Write response.
	h.render(w, "mods", map[string]interface{}{
		"Modules": mods,
	})
}

func (h *Handlers) Module(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse UUID.
	id, err := parseuuid(r.URL.Path, "/mod/")
	if err != nil {
		Error(w, err)
		return
	}

	// Fetch module.
	mod, err := h.srv.FindModuleByUUID(ctx, id)
	if err != nil {
		Error(w, err)
		return
	}

	pkgs, err := h.srv.ListModulePackages(ctx, mod)
	if err != nil {
		Error(w, err)
		return
	}

	// Write response.
	h.render(w, "mod", map[string]interface{}{
		"Module":   mod,
		"Packages": pkgs,
	})
}

func (h *Handlers) Package(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse UUID.
	id, err := parseuuid(r.URL.Path, "/pkg/")
	if err != nil {
		Error(w, err)
		return
	}

	// Fetch package.
	pkg, err := h.srv.FindPackageByUUID(ctx, id)
	if err != nil {
		Error(w, err)
		return
	}

	benchs, err := h.srv.ListPackageBenchmarks(ctx, pkg)
	if err != nil {
		Error(w, err)
		return
	}

	// Write response.
	h.render(w, "pkg", map[string]interface{}{
		"Package":    pkg,
		"Benchmarks": benchs,
	})
}

func (h *Handlers) Benchmark(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse UUID.
	id, err := parseuuid(r.URL.Path, "/bench/")
	if err != nil {
		Error(w, err)
		return
	}

	// Fetch benchmark.
	bench, err := h.srv.FindBenchmarkByUUID(ctx, id)
	if err != nil {
		Error(w, err)
		return
	}

	results, err := h.srv.ListBenchmarkResults(ctx, bench)
	if err != nil {
		Error(w, err)
		return
	}

	// Write response.
	h.render(w, "bench", map[string]interface{}{
		"Benchmark": bench,
		"Results":   results,
	})
}

func (h *Handlers) File(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse UUID.
	id, err := parseuuid(r.URL.Path, "/file/")
	if err != nil {
		Error(w, err)
		return
	}

	// Fetch file.
	file, err := h.srv.FindDataFileByUUID(ctx, id)
	if err != nil {
		Error(w, err)
		return
	}

	// Fetch raw data.
	rdr, err := h.datafs.Open(ctx, file.Name)
	if err != nil {
		Error(w, err)
		return
	}
	defer rdr.Close()

	type line struct {
		Num      int
		Contents string
	}
	var lines []line

	s := bufio.NewScanner(rdr)
	for s.Scan() {
		lines = append(lines, line{
			Num:      len(lines) + 1,
			Contents: s.Text(),
		})
	}
	if err := s.Err(); err != nil {
		Error(w, err)
		return
	}

	// Write response.
	h.render(w, "file", map[string]interface{}{
		"File":  file,
		"Lines": lines,
	})
}

func (h *Handlers) render(w http.ResponseWriter, name string, data interface{}) {
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		if err != nil {
			Error(w, err)
			return
		}
	}
}

func parseuuid(path, prefix string) (uuid.UUID, error) {
	if !strings.HasPrefix(path, prefix) {
		return uuid.Nil, fmt.Errorf("path %q expected to have prefix %q", path, prefix)
	}
	return uuid.Parse(path[len(prefix):])
}
