package main

import (
	"os/exec"
	"path/filepath"

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
func (r *Runner) Init() error {
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

	// Environment checks.
	r.GoExec("version")
	r.GoExec("env")

	return r.w.Error()
}

// Go builds a command with the downloaded go version.
func (r *Runner) Go(arg ...string) *exec.Cmd {
	return &exec.Cmd{
		Path: r.gobin,
		Args: append([]string{"go"}, arg...),
	}
}

// GoExec executes the go binary with the given arguments.
func (r *Runner) GoExec(arg ...string) {
	r.w.Exec(r.Go(arg...))
}

// Benchmark runs the benchmark job.
func (r *Runner) Benchmark(j Job) {
	defer lg.Scope(r.w, "benchmark")()

	r.w.Sandbox("bench")
	r.GoExec("mod", "init", "bench")
	r.GoExec("get", "-t", j.Module.String())
	r.GoExec(
		"test",
		"-run", "none^", // no tests
		"-bench", ".", // all benchmarks
		"-benchtime", "100ms", // 10ms each
		j.Module.Path+"/...",
	)
}
