package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/pkg/cfg"
	"github.com/mmcloughlin/cb/pkg/job"
	"github.com/mmcloughlin/cb/pkg/lg"
)

type Runner struct {
	w      *Workspace
	tc     Toolchain
	tuners []Tuner

	gobin    string
	wrappers []Wrapper
}

func NewRunner(w *Workspace, tc Toolchain) *Runner {
	return &Runner{
		w:  w,
		tc: tc,
	}
}

// Init initializes the runner.
func (r *Runner) Init(ctx context.Context) {
	defer lg.Scope(r.w, "initializing")()

	// Install toolchain.
	lg.Param(r.w, "toolchain", r.tc.String())
	goroot := r.w.Path("goroot")
	r.tc.Install(r.w, goroot)

	r.gobin = filepath.Join(goroot, "bin", "go")

	// Configure Go environment.
	r.w.SetEnv("GOROOT", goroot)
	r.w.SetEnv("GOPATH", r.w.EnsureDir("gopath"))
	r.w.SetEnv("GOCACHE", r.w.EnsureDir("gocache"))
	r.w.SetEnv("GO111MODULE", "on")
	r.w.DefineTool("AR", "ar")
	r.w.DefineTool("CC", "gcc")
	r.w.DefineTool("CXX", "g++")
	r.w.DefineTool("PKG_CONFIG", "pkg-config")

	// Environment checks.
	r.GoExec(ctx, "version")
	r.GoExec(ctx, "env")
}

// Clean up the runner.
func (r *Runner) Clean(ctx context.Context) {
	defer lg.Scope(r.w, "clean")()
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

// Tune applies the given tuning method to benchmark executions.
func (r *Runner) Tune(t Tuner) {
	r.tuners = append(r.tuners, t)
}

// Wrap configures the wrapper w to be applied to benchmark runs. Wrappers are applied in the order they are added.
func (r *Runner) Wrap(w ...Wrapper) {
	r.wrappers = append(r.wrappers, w...)
}

// Benchmark runs the benchmark suite.
func (r *Runner) Benchmark(ctx context.Context, s job.Suite) {
	defer lg.Scope(r.w, "benchmark")()

	// Apply tuners.
	for _, t := range r.tuners {
		if !t.Available() {
			continue
		}
		if err := t.Apply(); err != nil {
			r.w.seterr(err)
			return
		}
		defer func(t Tuner) {
			r.w.seterr(t.Reset())
		}(t)
	}

	// Setup.
	dir := r.w.Sandbox("bench")
	r.GoExec(ctx, "mod", "init", "bench")
	r.GoExec(ctx, "get", "-t", s.Module.String())

	// Open the output.
	outputfile := filepath.Join(dir, "bench.out")
	f, err := os.Create(outputfile)
	if err != nil {
		// TODO(mbm): cleaner seterr() mechanism on workspace
		r.w.seterr(err)
		return
	}
	defer f.Close()

	// Write static configuration.
	providers := cfg.Providers{
		ToolchainConfigurationProvider(r.tc),
	}

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
	filename := fmt.Sprintf("%s.out", uuid.New())
	path := filepath.Join(r.tc.String(), s.Module.String(), filename)
	r.w.Artifact(outputfile, path)
}

// testargs builds "go test" arguments for the given suite.
func testargs(s job.Suite) []string {
	args := []string{"test"}
	args = append(args, "-run", stringdefault(s.Tests, "."))
	if s.Short {
		args = append(args, "-short")
	}
	args = append(args, "-bench", stringdefault(s.Benchmarks, "."))
	args = append(args, "-benchtime", durationdefault(s.BenchTime, "1s"))
	args = append(args, "-timeout", durationdefault(s.Timeout, "0"))
	args = append(args, s.Module.Path+"/...")
	return args
}

// stringdefault returns s if non-empty, otherwise defaults to dflt.
func stringdefault(s, dflt string) string {
	if s != "" {
		return s
	}
	return dflt
}

// duration converts duration d to a string, using dflt if duration is 0.
func durationdefault(d time.Duration, dflt string) string {
	if d != 0 {
		return d.String()
	}
	return dflt
}
