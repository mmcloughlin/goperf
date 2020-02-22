package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/pkg/lg"
)

type Module struct {
	Path    string
	Version string
}

func (m Module) String() string {
	s := m.Path
	if m.Version != "" {
		s += "@" + m.Version
	}
	return s
}

type Job struct {
	Module Module
}

type Runner struct {
	w  *Workspace
	tc Toolchain

	gobin string
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

// Benchmark runs the benchmark job.
func (r *Runner) Benchmark(ctx context.Context, j Job) {
	defer lg.Scope(r.w, "benchmark")()

	// Setup.
	dir := r.w.Sandbox("bench")
	r.GoExec(ctx, "mod", "init", "bench")
	r.GoExec(ctx, "get", "-t", j.Module.String())

	// Run the benchmark.
	cmd := r.Go(
		ctx,
		"test",
		"-run", "none^", // no tests
		"-bench", ".", // all benchmarks
		"-benchtime", "10ms", // 10ms each
		j.Module.Path+"/...",
	)

	outputfile := filepath.Join(dir, "bench.out")
	f, err := os.Create(outputfile)
	if err != nil {
		// TODO(mbm): cleaner seterr() mechanism on workspace
		r.w.seterr(err)
		return
	}
	defer f.Close()

	cmd.Stdout = f
	r.w.Exec(cmd)

	// Save the result.
	filename := fmt.Sprintf("%s.out", uuid.New())
	path := filepath.Join(r.tc.String(), j.Module.String(), filename)
	r.w.Artifact(outputfile, path)
}
