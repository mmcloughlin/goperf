package main

import (
	"context"
	"flag"
	"path"
	"runtime"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/internal/flags"
	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/fs"
	"github.com/mmcloughlin/cb/pkg/runner"
)

const (
	owner = "klauspost"
	repo  = "compress"
	rev   = "b949da471e55fbe4393e6eb595602d936f5c312e"
)

type Run struct {
	command.Base

	wrappers []runner.Wrapper

	toolchainconfig flags.TypeParams
	output          string
	preserve        bool
}

func NewRun(b command.Base) *Run {
	return &Run{
		Base: b,

		toolchainconfig: flags.TypeParams{
			Type: "release",
			Params: flags.Params{
				{Key: "version", Value: runtime.Version()},
			},
		},
	}
}

// AddWrapper configures a wrapper around benchmark runs.
func (r *Run) AddWrapper(w runner.Wrapper) {
	r.wrappers = append(r.wrappers, w)
}

func (*Run) Name() string { return "run" }

func (*Run) Synopsis() string { return "run benchmark suites against a version of the go toolchain" }

func (*Run) Usage() string {
	return ""
}

func (cmd *Run) SetFlags(f *flag.FlagSet) {
	f.Var(&cmd.toolchainconfig, "toolchain", "toolchain configuration")
	f.StringVar(&cmd.output, "output", "", "output path")
	f.BoolVar(&cmd.preserve, "preserve", false, "preserve working directory")
}

func (cmd *Run) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	// Build toolchain.
	tc, err := runner.NewToolchain(cmd.toolchainconfig.Type, cmd.toolchainconfig.Params.Map())
	if err != nil {
		return cmd.Error(err)
	}

	// Construct workspace.
	w, err := runner.NewWorkspace(runner.WithLogger(cmd.Log))
	if err != nil {
		return cmd.Error(err)
	}

	if cmd.output != "" {
		w.Options(runner.WithArtifactStore(fs.NewLocal(cmd.output)))
	}

	// Initialize runner.
	r := runner.NewRunner(w, tc)
	r.Wrap(cmd.wrappers...)

	r.Init(ctx)

	// Run benchmark.
	mod := runner.Module{
		Path:    path.Join("github.com", owner, repo),
		Version: rev,
	}
	job := runner.Job{
		Module: mod,
	}
	r.Benchmark(ctx, job)

	// Clean.
	if !cmd.preserve {
		r.Clean(ctx)
	}

	return cmd.Status(w.Error())
}