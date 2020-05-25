package dashboard

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
	analysis "golang.org/x/perf/analysis/app"

	"github.com/mmcloughlin/goperf/app/brand"
	"github.com/mmcloughlin/goperf/app/change"
	"github.com/mmcloughlin/goperf/app/db"
	"github.com/mmcloughlin/goperf/app/entity"
	"github.com/mmcloughlin/goperf/app/env"
	"github.com/mmcloughlin/goperf/app/httputil"
	"github.com/mmcloughlin/goperf/internal/errutil"
	"github.com/mmcloughlin/goperf/pkg/fs"
	"github.com/mmcloughlin/goperf/pkg/units"
)

type Handlers struct {
	db       *db.DB
	staticfs fs.Readable
	datafs   fs.Readable
	cc       httputil.CacheControl

	mux       *http.ServeMux
	static    *httputil.Static
	templates *Templates
	envcache  sync.Map
	log       *zap.Logger
}

type Option func(*Handlers)

func WithTemplates(t *Templates) Option {
	return func(h *Handlers) { h.templates = t }
}

func WithStaticFileSystem(r fs.Readable) Option {
	return func(h *Handlers) { h.staticfs = r }
}

func WithDataFileSystem(r fs.Readable) Option {
	return func(h *Handlers) { h.datafs = r }
}

func WithCacheControl(cc httputil.CacheControl) Option {
	return func(h *Handlers) { h.cc = cc }
}

func WithLogger(l *zap.Logger) Option {
	return func(h *Handlers) { h.log = l.Named("handlers") }
}

func NewHandlers(d *db.DB, opts ...Option) *Handlers {
	// Configure.
	h := &Handlers{
		db:        d,
		staticfs:  StaticFileSystem,
		datafs:    fs.Null,
		cc:        httputil.CacheControlNever,
		mux:       http.NewServeMux(),
		templates: NewTemplates(TemplateFileSystem),
		log:       zap.NewNop(),
	}
	for _, opt := range opts {
		opt(h)
	}

	// Setup mux.
	h.mux.Handle("/", h.handlerFunc(h.Index))
	h.mux.Handle("/mods/", h.handlerFunc(h.Modules))
	h.mux.Handle("/mod/", h.handlerFunc(h.Module))
	h.mux.Handle("/pkg/", h.handlerFunc(h.Package))
	h.mux.Handle("/bench/", h.handlerFunc(h.Benchmark))
	h.mux.Handle("/result/", h.handlerFunc(h.Result))
	h.mux.Handle("/file/", h.handlerFunc(h.File))
	h.mux.Handle("/commit/", h.handlerFunc(h.Commit))
	h.mux.Handle("/chgs/", h.handlerFunc(h.Changes))

	h.mux.Handle("/about/", h.handlerFunc(h.About))

	// Static assets.
	h.static = httputil.NewStatic(h.staticfs)
	h.static.SetLogger(h.log)
	h.mux.Handle("/static/", http.StripPrefix("/static/", h.handler(h.static)))

	return h
}

func (h *Handlers) handler(handler httputil.Handler) http.Handler {
	return httputil.CacheHandler(h.cc, httputil.ErrorHandler{
		Handler: handler,
		Log:     h.log,
	})
}

func (h *Handlers) handlerFunc(handler httputil.HandlerFunc) http.Handler { return h.handler(handler) }

func (h *Handlers) Init(ctx context.Context) error {
	// Template function for static paths.
	h.templates.Func("static", func(name string) (string, error) {
		p, err := h.static.Path(ctx, name)
		if err != nil {
			return "", err
		}
		return path.Join("/static/", p), nil
	})

	// Color scheme.
	h.templates.Func("color", brand.Color)

	// Linkify.
	h.templates.Func("linkify", linkify)

	return h.templates.Init(ctx)
}

func (h *Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handlers) Index(w http.ResponseWriter, r *http.Request) error {
	bymaxpercent := func(g *CommitChangeGroup) float64 {
		return -g.MaxAbsPercentChange()
	}
	return h.changes(w, r, "index", db.ChangeFilter{
		MinEffectSize:             20,
		MaxRankByAbsPercentChange: 3,
	}, bymaxpercent)
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
		"Package":         pkg,
		"BenchmarkGroups": benchmarkGroups(benchs),
	})
}

type BenchmarkGroup struct {
	Name  string
	Units []*entity.Benchmark
}

func benchmarkGroups(benchs []*entity.Benchmark) []*BenchmarkGroup {
	// Group by name.
	byname := map[string]*BenchmarkGroup{}
	for _, bench := range benchs {
		name := bench.FullName
		if _, ok := byname[name]; !ok {
			byname[name] = &BenchmarkGroup{Name: name}
		}
		byname[name].Units = append(byname[name].Units, bench)
	}

	// Convert to list.
	groups := []*BenchmarkGroup{}
	for _, group := range byname {
		groups = append(groups, group)
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Name < groups[j].Name
	})

	for _, group := range groups {
		u := group.Units
		sort.Slice(u, func(i, j int) bool {
			return units.Less(u[i].Unit, u[j].Unit)
		})
	}

	return groups
}

func (h *Handlers) Benchmark(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Parse UUID.
	id, err := parseuuid(r.URL.Path, "/bench/")
	if err != nil {
		return err
	}

	// Determine commit range.
	cr, err := h.commitRange(r)
	if err != nil {
		return err
	}

	// Fetch benchmark.
	bench, err := h.db.FindBenchmarkByUUID(ctx, id)
	if err != nil {
		return err
	}

	points, err := h.db.ListBenchmarkPoints(ctx, bench, cr)
	if err != nil {
		return err
	}

	// Group by environment.
	groups, err := h.groups(ctx, points, bench.Unit)
	if err != nil {
		return err
	}

	// Write response.
	return h.render(ctx, w, "bench", map[string]interface{}{
		"Benchmark":        bench,
		"CommitIndexRange": cr,
		"PointsGroups":     groups,
	})
}

// PointsGroup is a benchmark timeseries for a given environment.
type PointsGroup struct {
	Title       string
	Environment entity.Properties
	Points      entity.Points
	Filtered    []float64
	Quantities  []units.Quantity
}

func (h *Handlers) groups(ctx context.Context, points entity.Points, unit string) ([]*PointsGroup, error) {
	// Group by environment.
	byenv := map[uuid.UUID]entity.Points{}
	for _, point := range points {
		byenv[point.EnvironmentUUID] = append(byenv[point.EnvironmentUUID], point)
	}

	// Fetch environment objects and build groups.
	groups := []*PointsGroup{}
	for id, points := range byenv {
		e, err := h.env(ctx, id)
		if err != nil {
			return nil, err
		}
		groups = append(groups, &PointsGroup{
			Title:       env.Title(e),
			Environment: e,
			Points:      points,
		})
	}

	// Apply KZA filtering.
	for _, group := range groups {
		group.Filtered = analysis.AdaptiveKolmogorovZurbenko(group.Points.Values(), 31, 5)
	}

	// Convert to quantity.
	for _, group := range groups {
		for _, p := range group.Points {
			q := units.Quantity{Value: p.Value, Unit: unit}
			group.Quantities = append(group.Quantities, units.Humanize(q))
		}
	}

	// Sort by descending size.
	sort.Slice(groups, func(i, j int) bool {
		return len(groups[i].Points) > len(groups[j].Points)
	})

	return groups, nil
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

func (h *Handlers) File(w http.ResponseWriter, r *http.Request) (err error) {
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
	defer errutil.CheckClose(&err, rdr)

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

	// Attempt to find index.
	idx, err := h.db.FindCommitIndexBySHA(ctx, sha)
	if err != nil {
		idx = -1
	}

	// Find associated changes, if we have a commit index. Use 0 as effect size
	// threshold to return everything.
	var changes []*Change
	if idx >= 0 {
		changes, err = h.changesForCommit(ctx, idx)
		if err != nil {
			return err
		}
	}

	// Write response.
	return h.render(ctx, w, "commit", map[string]interface{}{
		"Commit":      commit,
		"CommitIndex": idx,
		"Changes":     changes,
	})
}

func (h *Handlers) changesForCommit(ctx context.Context, idx int) ([]*Change, error) {
	chgs, err := h.db.ListChangeSummariesForCommitIndex(ctx, idx, db.ChangeFilter{
		MinEffectSize: 0, // everything
	})
	if err != nil {
		return nil, err
	}

	if len(chgs) == 0 {
		return nil, nil
	}

	groups, err := h.groupChanges(ctx, chgs)
	if err != nil {
		return nil, err
	}

	if len(groups) != 1 {
		return nil, errutil.AssertionFailure("expected one change group")
	}

	return groups[0].Changes, nil
}

func (h *Handlers) Changes(w http.ResponseWriter, r *http.Request) error {
	byindex := func(g *CommitChangeGroup) float64 {
		return -float64(g.Index)
	}
	return h.changes(w, r, "chgs", db.ChangeFilter{
		MinEffectSize:       10,
		MaxRankByEffectSize: 5,
	}, byindex)
}

func (h *Handlers) changes(w http.ResponseWriter, r *http.Request, tmpl string, filter db.ChangeFilter, sortby func(*CommitChangeGroup) float64) error {
	ctx := r.Context()

	// Determine commit range.
	cr, err := h.commitRange(r)
	if err != nil {
		return err
	}

	// Fetch changes.
	chgs, err := h.db.ListChangeSummaries(ctx, cr, filter)
	if err != nil {
		return err
	}

	groups, err := h.groupChanges(ctx, chgs)
	if err != nil {
		return err
	}

	sort.Slice(groups, func(i, j int) bool {
		return sortby(groups[i]) < sortby(groups[j])
	})

	// Write response.
	return h.render(ctx, w, tmpl, map[string]interface{}{
		"CommitChangeGroups": groups,
	})
}

// CommitChangeGroup is a group of significant changes for a commit.
type CommitChangeGroup struct {
	Index   int
	SHA     string
	Subject string
	Changes []*Change
}

func (g *CommitChangeGroup) MaxAbsPercentChange() float64 {
	max := math.Inf(-1)
	for _, c := range g.Changes {
		max = math.Max(max, math.Abs(c.Percent()))
	}
	return max
}

// Change is a single change.
type Change struct {
	Benchmark   *entity.Benchmark
	Environment string
	change.Change
}

func (c *Change) Type() change.Type {
	return change.Classify(c.Pre.Mean, c.Post.Mean, c.Benchmark.Unit)
}

func (h *Handlers) groupChanges(ctx context.Context, cs []*entity.ChangeSummary) ([]*CommitChangeGroup, error) {
	// Group by commit.
	byidx := map[int]*CommitChangeGroup{}
	for _, c := range cs {
		if byidx[c.CommitIndex] == nil {
			byidx[c.CommitIndex] = &CommitChangeGroup{
				Index:   c.CommitIndex,
				SHA:     c.CommitSHA,
				Subject: c.CommitSubject,
			}
		}

		e, err := h.env(ctx, c.EnvironmentUUID)
		if err != nil {
			return nil, err
		}

		byidx := byidx[c.CommitIndex]
		byidx.Changes = append(byidx.Changes, &Change{
			Benchmark:   c.Benchmark,
			Environment: env.Short(e),
			Change:      c.Change,
		})
	}

	// Convert to a list.
	groups := make([]*CommitChangeGroup, 0, len(byidx))
	for _, group := range byidx {
		groups = append(groups, group)
	}

	// Most recent commits first.
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Index > groups[j].Index
	})

	// Within each group, sort on effect size.
	for _, group := range byidx {
		chgs := group.Changes
		sort.Slice(chgs, func(i, j int) bool {
			return math.Abs(chgs[i].EffectSize) > math.Abs(chgs[j].EffectSize)
		})
	}

	return groups, nil
}

// commitRange determines a specified commit range for the given request.
func (h *Handlers) commitRange(r *http.Request) (entity.CommitIndexRange, error) {
	ctx := r.Context()

	// Look for explicit min/max.
	min := intparam(r, "min", -1)
	max := intparam(r, "max", -1)
	if 0 <= min && min < max {
		return entity.CommitIndexRange{
			Min: min,
			Max: max,
		}, nil
	}

	// Support center "c" and total number "n".
	n := intparam(r, "n", 300)
	mid := intparam(r, "c", -1)

	if mid >= 0 {
		return entity.CommitIndexRange{
			Min: mid - n/2,
			Max: mid + n/2,
		}, nil
	}

	// Determine commit index.
	idx, err := h.db.MostRecentCommitIndex(ctx)
	if err != nil {
		return entity.CommitIndexRange{}, err
	}

	return entity.CommitIndexRange{
		Min: idx - n,
		Max: idx,
	}, nil
}

func (h *Handlers) About(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	return h.render(ctx, w, "about", nil)
}

func (h *Handlers) render(ctx context.Context, w io.Writer, name string, data interface{}) error {
	return h.templates.ExecuteTemplate(ctx, w, name+".gohtml", "main", data)
}

// env looks up the environment with the given ID, returning from a cache if it's been accessed before.
func (h *Handlers) env(ctx context.Context, id uuid.UUID) (entity.Properties, error) {
	v, ok := h.envcache.Load(id)
	if ok {
		return v.(entity.Properties), nil
	}

	e, err := h.db.FindPropertiesByUUID(ctx, id)
	if err != nil {
		return nil, err
	}

	h.envcache.Store(id, e)
	return e, nil
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

func intparam(r *http.Request, key string, dflt int) int {
	v, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil {
		return dflt
	}
	return v
}
