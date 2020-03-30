package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	analysis "golang.org/x/perf/analysis/app"

	"github.com/mmcloughlin/cb/app/brand"
	"github.com/mmcloughlin/cb/app/db"
	"github.com/mmcloughlin/cb/pkg/fs"
	"github.com/mmcloughlin/cb/pkg/units"
)

type Handlers struct {
	db     *db.DB
	static fs.Readable
	datafs fs.Readable

	mux       *http.ServeMux
	templates *Templates
}

type Option func(*Handlers)

func WithTemplates(t *Templates) Option {
	return func(h *Handlers) { h.templates = t }
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
		db:        d,
		static:    StaticFileSystem,
		datafs:    fs.Null,
		mux:       http.NewServeMux(),
		templates: NewTemplates(TemplateFileSystem),
	}
	for _, opt := range opts {
		opt(h)
	}

	// Setup mux.
	h.mux.HandleFunc("/mods/", h.Modules)
	h.mux.HandleFunc("/mod/", h.Module)
	h.mux.HandleFunc("/pkg/", h.Package)
	h.mux.HandleFunc("/bench/", h.Benchmark)
	h.mux.HandleFunc("/result/", h.Result)
	h.mux.HandleFunc("/file/", h.File)
	h.mux.HandleFunc("/commit/", h.Commit)

	// Static assets.
	static := NewStatic(h.static)
	h.mux.Handle("/static/", http.StripPrefix("/static/", static))

	h.mux.Handle("/favicon.ico", ProxySingleURL(&url.URL{Scheme: "https", Host: "golang.org", Path: "/favicon.ico"}))

	return h
}

func (h *Handlers) Init(ctx context.Context) error {
	h.templates.Func("color", brand.Color)
	return h.templates.Init(ctx)
}

func (h *Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handlers) Modules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Fetch modules.
	mods, err := h.db.ListModules(ctx)
	if err != nil {
		Error(w, err)
		return
	}

	// Write response.
	h.render(ctx, w, "mods", map[string]interface{}{
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
	mod, err := h.db.FindModuleByUUID(ctx, id)
	if err != nil {
		Error(w, err)
		return
	}

	pkgs, err := h.db.ListModulePackages(ctx, mod)
	if err != nil {
		Error(w, err)
		return
	}

	// Write response.
	h.render(ctx, w, "mod", map[string]interface{}{
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
	pkg, err := h.db.FindPackageByUUID(ctx, id)
	if err != nil {
		Error(w, err)
		return
	}

	benchs, err := h.db.ListPackageBenchmarks(ctx, pkg)
	if err != nil {
		Error(w, err)
		return
	}

	// Write response.
	h.render(ctx, w, "pkg", map[string]interface{}{
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
	bench, err := h.db.FindBenchmarkByUUID(ctx, id)
	if err != nil {
		Error(w, err)
		return
	}

	points, err := h.db.ListBenchmarkPoints(ctx, bench, 256)
	if err != nil {
		Error(w, err)
		return
	}

	// Apply KZA filter.
	kza := analysis.AdaptiveKolmogorovZurbenko(points.Values(), 31, 5)

	// Write response.
	h.render(ctx, w, "bench", map[string]interface{}{
		"Benchmark": bench,
		"Points":    points,
		"Filtered":  kza,
	})
}

func (h *Handlers) Result(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse UUID.
	id, err := parseuuid(r.URL.Path, "/result/")
	if err != nil {
		Error(w, err)
		return
	}

	// Fetch result.
	result, err := h.db.FindResultByUUID(ctx, id)
	if err != nil {
		Error(w, err)
		return
	}

	quantity := units.Humanize(units.Quantity{
		Value: result.Value,
		Unit:  result.Benchmark.Unit,
	})

	// Write response.
	h.render(ctx, w, "result", map[string]interface{}{
		"Result":   result,
		"Quantity": quantity,
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
	file, err := h.db.FindDataFileByUUID(ctx, id)
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
	h.render(ctx, w, "file", map[string]interface{}{
		"File":  file,
		"Lines": lines,
	})
}

func (h *Handlers) Commit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract commit SHA.
	sha, err := stripprefix(r.URL.Path, "/commit/")
	if err != nil {
		Error(w, err)
		return
	}

	// Fetch commit.
	commit, err := h.db.FindCommitBySHA(ctx, sha)
	if err != nil {
		Error(w, err)
		return
	}

	// Write response.
	h.render(ctx, w, "commit", map[string]interface{}{
		"Commit": commit,
	})
}

func (h *Handlers) render(ctx context.Context, w http.ResponseWriter, name string, data interface{}) {
	if err := h.templates.ExecuteTemplate(ctx, w, name+".gohtml", "main", data); err != nil {
		Error(w, err)
		return
	}
}

func parseuuid(path, prefix string) (uuid.UUID, error) {
	rest, err := stripprefix(path, prefix)
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(rest)
}

func stripprefix(path, prefix string) (string, error) {
	if !strings.HasPrefix(path, prefix) {
		return "", fmt.Errorf("path %q expected to have prefix %q", path, prefix)
	}
	return path[len(prefix):], nil
}
