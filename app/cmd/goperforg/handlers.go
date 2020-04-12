package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	analysis "golang.org/x/perf/analysis/app"

	"github.com/mmcloughlin/cb/app/brand"
	"github.com/mmcloughlin/cb/app/db"
	"github.com/mmcloughlin/cb/app/httputil"
	"github.com/mmcloughlin/cb/pkg/fs"
	"github.com/mmcloughlin/cb/pkg/units"
)

type Handlers struct {
	db     *db.DB
	static fs.Readable
	datafs fs.Readable

	mux       *http.ServeMux
	templates *Templates
	log       *zap.Logger
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

func WithLogger(l *zap.Logger) Option {
	return func(h *Handlers) { h.log = l.Named("handlers") }
}

func NewHandlers(d *db.DB, opts ...Option) *Handlers {
	// Configure.
	h := &Handlers{
		db:        d,
		static:    StaticFileSystem,
		datafs:    fs.Null,
		mux:       http.NewServeMux(),
		templates: NewTemplates(TemplateFileSystem),
		log:       zap.NewNop(),
	}
	for _, opt := range opts {
		opt(h)
	}

	// Setup mux.
	h.mux.Handle("/mods/", h.handlerFunc(h.Modules))
	h.mux.Handle("/mod/", h.handlerFunc(h.Module))
	h.mux.Handle("/pkg/", h.handlerFunc(h.Package))
	h.mux.Handle("/bench/", h.handlerFunc(h.Benchmark))
	h.mux.Handle("/result/", h.handlerFunc(h.Result))
	h.mux.Handle("/file/", h.handlerFunc(h.File))
	h.mux.Handle("/commit/", h.handlerFunc(h.Commit))

	// Static assets.
	static := h.handler(httputil.NewStatic(h.static))
	h.mux.Handle("/static/", http.StripPrefix("/static/", static))

	iconproxy := httputil.ProxySingleURL(&url.URL{Scheme: "https", Host: "golang.org", Path: "/favicon.ico"})
	h.mux.Handle("/favicon.ico", h.handler(iconproxy))

	return h
}

func (h *Handlers) handler(handler httputil.Handler) http.Handler {
	return httputil.ErrorHandler{
		Handler: handler,
		Log:     h.log,
	}
}

func (h *Handlers) handlerFunc(handler httputil.HandlerFunc) http.Handler { return h.handler(handler) }

func (h *Handlers) Init(ctx context.Context) error {
	h.templates.Func("color", brand.Color)
	return h.templates.Init(ctx)
}

func (h *Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handlers) Modules(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Fetch modules.
	mods, err := h.db.ListModules(ctx)
	if err != nil {
		return err
	}

	// Write response.
	return h.render(ctx, w, "mods", map[string]interface{}{
		"Modules": mods,
	})
}

func (h *Handlers) Module(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Parse UUID.
	id, err := parseuuid(r.URL.Path, "/mod/")
	if err != nil {
		return err
	}

	// Fetch module.
	mod, err := h.db.FindModuleByUUID(ctx, id)
	if err != nil {
		return err
	}

	pkgs, err := h.db.ListModulePackages(ctx, mod)
	if err != nil {
		return err
	}

	// Write response.
	return h.render(ctx, w, "mod", map[string]interface{}{
		"Module":   mod,
		"Packages": pkgs,
	})
}

func (h *Handlers) Package(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Parse UUID.
	id, err := parseuuid(r.URL.Path, "/pkg/")
	if err != nil {
		return err
	}

	// Fetch package.
	pkg, err := h.db.FindPackageByUUID(ctx, id)
	if err != nil {
		return err
	}

	benchs, err := h.db.ListPackageBenchmarks(ctx, pkg)
	if err != nil {
		return err
	}

	// Write response.
	return h.render(ctx, w, "pkg", map[string]interface{}{
		"Package":    pkg,
		"Benchmarks": benchs,
	})
}

func (h *Handlers) Benchmark(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Parse UUID.
	id, err := parseuuid(r.URL.Path, "/bench/")
	if err != nil {
		return err
	}

	// Fetch benchmark.
	bench, err := h.db.FindBenchmarkByUUID(ctx, id)
	if err != nil {
		return err
	}

	points, err := h.db.ListBenchmarkPoints(ctx, bench, 256)
	if err != nil {
		return err
	}

	// Apply KZA filter.
	kza := analysis.AdaptiveKolmogorovZurbenko(points.Values(), 31, 5)

	// Write response.
	return h.render(ctx, w, "bench", map[string]interface{}{
		"Benchmark": bench,
		"Points":    points,
		"Filtered":  kza,
	})
}

func (h *Handlers) Result(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Parse UUID.
	id, err := parseuuid(r.URL.Path, "/result/")
	if err != nil {
		return err
	}

	// Fetch result.
	result, err := h.db.FindResultByUUID(ctx, id)
	if err != nil {
		return err
	}

	quantity := units.Humanize(units.Quantity{
		Value: result.Value,
		Unit:  result.Benchmark.Unit,
	})

	// Write response.
	return h.render(ctx, w, "result", map[string]interface{}{
		"Result":   result,
		"Quantity": quantity,
	})
}

func (h *Handlers) File(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Parse UUID.
	id, err := parseuuid(r.URL.Path, "/file/")
	if err != nil {
		return err
	}

	// Was a line selection specified?
	hl := 0
	if ln, err := strconv.Atoi(r.URL.Query().Get("hl")); err == nil {
		hl = ln
	}

	// Fetch file.
	file, err := h.db.FindDataFileByUUID(ctx, id)
	if err != nil {
		return err
	}

	// Fetch raw data.
	rdr, err := h.datafs.Open(ctx, file.Name)
	if err != nil {
		return err
	}
	defer rdr.Close()

	type line struct {
		Num       int
		Contents  string
		Highlight bool
	}
	var lines []line

	s := bufio.NewScanner(rdr)
	for s.Scan() {
		n := len(lines) + 1
		lines = append(lines, line{
			Num:       n,
			Contents:  s.Text(),
			Highlight: n == hl,
		})
	}
	if err := s.Err(); err != nil {
		return err
	}

	// Write response.
	return h.render(ctx, w, "file", map[string]interface{}{
		"File":  file,
		"Lines": lines,
	})
}

func (h *Handlers) Commit(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Extract commit SHA.
	sha, err := stripprefix(r.URL.Path, "/commit/")
	if err != nil {
		return err
	}

	// Fetch commit.
	commit, err := h.db.FindCommitBySHA(ctx, sha)
	if err != nil {
		return err
	}

	// Write response.
	return h.render(ctx, w, "commit", map[string]interface{}{
		"Commit": commit,
	})
}

func (h *Handlers) render(ctx context.Context, w io.Writer, name string, data interface{}) error {
	return h.templates.ExecuteTemplate(ctx, w, name+".gohtml", "main", data)
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
