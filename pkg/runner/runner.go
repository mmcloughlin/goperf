// Package runner implements sandboxed Go benchmark execution.
package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/mmcloughlin/goperf/pkg/cfg"
	"github.com/mmcloughlin/goperf/pkg/job"
	"github.com/mmcloughlin/goperf/pkg/lg"
)

type Runner struct {
	w         *Workspace
	tc        Toolchain
	goproxy   string
	tuners    []Tuner
	providers cfg.Providers

	gobin    string
	wrappers []Wrapper
}

func NewRunner(w *Workspace, tc Toolchain) *Runner {
	return &Runner{
		w:       w,
		tc:      tc,
		goproxy: "https://proxy.golang.org",
	}
}

// SetGoProxy sets the GOPROXY environment variable.
func (r *Runner) SetGoProxy(proxy string) {
	r.w.Log.Info("set go proxy", zap.String("GOPROXY", proxy))
	r.goproxy = proxy
}

// Init initializes the runner.
func (r *Runner) Init(ctx context.Context) {
	defer lg.Scope(r.w.Log, "initializing")()

	// Install toolchain.
	r.w.Log.Info("install toolchain", zap.Stringer("toolchain", r.tc))
	goroot := r.w.Path("goroot")
	r.tc.Install(r.w, goroot)

	gorootbin := filepath.Join(goroot, "bin")
	r.gobin = filepath.Join(gorootbin, "go")

	// Configure Go environment.
	r.w.SetEnv("GOROOT", goroot)
	r.w.SetEnv("GOPATH", r.w.EnsureDir("gopath"))
	r.w.SetEnv("GOCACHE", r.w.EnsureDir("gocache"))
	r.w.SetEnv("GO111MODULE", "on")
	r.w.SetEnv("GOPROXY", r.goproxy)
	r.w.DefineTool("AR", "ar")
	r.w.DefineTool("CC", "gcc")
	r.w.DefineTool("CXX", "g++")
	r.w.DefineTool("PKG_CONFIG", "pkg-config")

	// Environment required by standard library tests.
	// BenchmarkExecHostname calls "hostname".
	// https://github.com/golang/go/blob/83610c90bbe4f5f0b18ac01da3f3921c2f7090e4/src/os/exec/bench_test.go#L11
	r.w.ExposeTool("hostname")

	// Environment checks.
	r.GoExec(ctx, "version")
	r.GoExec(ctx, "env")
}

// Clean up the runner.
func (r *Runner) Clean(ctx context.Context) {
	defer lg.Scope(r.w.Log, "clean")()
	r.GoExec(ctx, "clean", "-cache", "-testcache", "-modcache")
	r.w.Clean()
}

// Go builds a command with the downloaded go version.
func (r *Runner) Go(ctx context.Context, arg ...string) *exec.Cmd {
	return exec.CommandContext(ctx, r.gobin, arg...)
}

// GoExec executes the go binary with the given arguments.
func (r *Runner) GoExec(ctx context.Context, arg ...string) {
	r.w.Exec(r.Go(ctx, arg...))
}

// AddConfigurationProvider adds a configuration provider that will be applied
// to every benchmark run.
func (r *Runner) AddConfigurationProvider(p cfg.Provider) {
	r.providers = append(r.providers, p)
}

// Tune applies the given tuning method to benchmark executions.
func (r *Runner) Tune(t Tuner) {
	r.tuners = append(r.tuners, t)
}

// Wrap configures the wrapper w to be applied to benchmark runs. Wrappers are applied in the order they are added.
func (r *Runner) Wrap(w ...Wrapper) {
	r.wrappers = append(r.wrappers, w...)
}

// Benchmark runs the benchmark suite.
func (r *Runner) Benchmark(ctx context.Context, s job.Suite, output string) {
	defer lg.Scope(r.w.Log, "benchmark")()

	// Apply tuners.
	for _, t := range r.tuners {
		log := r.w.Log.With(zap.String("tuner", t.Name()))
		if !t.Available() {
			log.Info("tuner unavailable")
			continue
		}
		log.Info("applying tuner")
		if err := t.Apply(); err != nil {
			r.w.seterr(err)
			return
		}
		defer func(t Tuner) {
			log.Info("reset tuner")
			r.w.seterr(t.Reset())
		}(t)
	}

	// Setup.
	dir := r.w.Sandbox("bench")
	r.GoExec(ctx, "mod", "init", "bench")

	if !s.Module.IsMeta() {
		r.GoExec(ctx, "get", "-t", s.Module.String())
	}

	// Open the output.
	outputfile := filepath.Join(dir, "bench.out")
	f, err := os.Create(outputfile)
	if err != nil {
		// TODO(mbm): cleaner seterr() mechanism on workspace
		r.w.seterr(err)
		return
	}

	// Write static configuration.
	providers := cfg.Providers{
		ToolchainConfigurationProvider(r.tc),
		suiteconfig(s),
	}
	providers = append(providers, r.providers...)

	c, err := providers.Configuration()
	if err != nil {
		r.w.seterr(err)
		return
	}

	if err := cfg.Write(f, c); err != nil {
		r.w.seterr(err)
		return
	}

	// Prepare the benchmark.
	args := testargs(s)
	cmd := r.Go(ctx, args...)

	for _, w := range r.wrappers {
		w(cmd)
	}

	cmd.Stdout = f
	r.w.Exec(cmd)

	// Save the result.
	r.w.Artifact(outputfile, output)
}

func suiteconfig(s job.Suite) cfg.Provider {
	return cfg.Section(
		"suite",
		"benchmark suite metadata",
		cfg.Property("mod", "benchmark suite module", s.Module),
		cfg.Property("modpath", "module path for the benchmark suite", cfg.StringValue(s.Module.Path)),
		cfg.Property("modversion", "module version for the benchmark suite", cfg.StringValue(s.Module.Version)),
		cfg.Property("benchmarks", "benchmarks regular expression", cfg.StringValue(s.BenchmarkRegex())),
		cfg.Property("benchtime", "minimum benchmark time", s.BenchmarkTime()),
		cfg.Property("tests", "tests regular expression", cfg.StringValue(s.TestRegex())),
		cfg.Property("short", "short test mode enabled", cfg.BoolValue(s.Short)),
		cfg.Property("timeout", "timeout for total test binary execution time", s.Timeout),
	)
}

// testargs builds "go test" arguments for the given suite.
func testargs(s job.Suite) []string {
	if s.Module.IsMeta() {
		s.Tests = job.SkipTests
	}
	args := []string{"test"}
	args = append(args, "-run", s.TestRegex())
	if s.Short {
		args = append(args, "-short")
	}
	args = append(args, "-bench", s.BenchmarkRegex())
	args = append(args, "-benchtime", s.BenchmarkTime().String())
	args = append(args, "-timeout", durationdefault(s.Timeout, "0"))
	if s.Module.IsMeta() {
		args = append(args, s.Module.Path)
	} else {
		args = append(args, s.Module.Path+"/...")
	}
	return args
}

// duration converts duration d to a string, using dflt if duration is 0.
func durationdefault(d time.Duration, dflt string) string {
	if d != 0 {
		return d.String()
	}
	return dflt
}
