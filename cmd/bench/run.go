package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"
	"time"

	"github.com/google/subcommands"
	"github.com/google/uuid"

	"github.com/mmcloughlin/goperf/internal/flags"
	"github.com/mmcloughlin/goperf/pkg/command"
	"github.com/mmcloughlin/goperf/pkg/fs"
	"github.com/mmcloughlin/goperf/pkg/job"
	"github.com/mmcloughlin/goperf/pkg/platform"
	"github.com/mmcloughlin/goperf/pkg/runner"
)

var defaultModule = job.Module{
	Path:    "github.com/klauspost/compress",
	Version: "b949da471e55fbe4393e6eb595602d936f5c312e",
}

type Run struct {
	command.Base
	*platform.Platform

	toolchainconfig flags.TypeParams
	mod             job.Module

	output   string
	preserve bool
	goproxy  string
}

func NewRun(b command.Base, p *platform.Platform) *Run {
	return &Run{
		Base:     b,
		Platform: p,

		toolchainconfig: flags.TypeParams{
			Type: "release",
			Params: flags.Params{
				{Key: "version", Value: runtime.Version()},
			},
		},
	}
}

func (*Run) Name() string { return "run" }

func (*Run) Synopsis() string { return "run benchmark suites against a version of the go toolchain" }

func (*Run) Usage() string {
	return ""
}

func (cmd *Run) SetFlags(f *flag.FlagSet) {
	f.Var(&cmd.toolchainconfig, "toolchain", "toolchain configuration")
	f.StringVar(&cmd.mod.Path, "modpath", defaultModule.Path, "module path")
	f.StringVar(&cmd.mod.Version, "modrev", defaultModule.Version, "module revision")

	f.StringVar(&cmd.output, "output", "", "output path")
	f.BoolVar(&cmd.preserve, "preserve", false, "preserve working directory")
	f.StringVar(&cmd.goproxy, "goproxy", "", "GOPROXY value for the benchark runner")

	cmd.Platform.SetFlags(f)
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

	if cmd.goproxy != "" {
		r.SetGoProxy(cmd.goproxy)
	}

	if err := cmd.ConfigureRunner(r); err != nil {
		return cmd.Error(err)
	}

	r.Init(ctx)

	// Run benchmark.
	suite := job.Suite{
		Module:    cmd.mod,
		Short:     true,
		BenchTime: 10 * time.Millisecond,
	}
	output := fmt.Sprintf("%s.out", uuid.New())
	r.Benchmark(ctx, suite, output)

	// Clean.
	if !cmd.preserve {
		r.Clean(ctx)
	}

	return cmd.Status(w.Error())
}
