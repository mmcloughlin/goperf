package main

import (
	"context"
	"flag"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/app/consumer"
	"github.com/mmcloughlin/cb/app/gcs"
	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/job"
	"github.com/mmcloughlin/cb/pkg/lg"
	"github.com/mmcloughlin/cb/pkg/platform"
	"github.com/mmcloughlin/cb/pkg/runner"
)

var (
	// TODO(mbm): remove hardcoded subscription
	subscription = "projects/contbench/subscriptions/worker_jobs"
	// TODO(mbm): remove hardcoded bucket
	bucket = "contbench_results"
)

type Run struct {
	command.Base
	*platform.Platform
}

func NewRun(b command.Base, p *platform.Platform) *Run {
	return &Run{
		Base:     b,
		Platform: p,
	}
}

func (*Run) Name() string { return "run" }

func (*Run) Synopsis() string {
	return "run benchmark job subscriber and executor"
}

func (*Run) Usage() string {
	return ""
}

func (cmd *Run) SetFlags(f *flag.FlagSet) {
	cmd.Platform.SetFlags(f)
}

func (cmd *Run) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	h := &Handler{
		Logger:   cmd.Log,
		Platform: cmd.Platform,
	}
	c, err := consumer.New(ctx, subscription, h, consumer.WithLogger(cmd.Log))
	if err != nil {
		return cmd.Error(err)
	}
	defer c.Close()
	return cmd.Status(c.Receive(ctx))
}

type Handler struct {
	lg.Logger
	*platform.Platform
}

func (h *Handler) Handle(ctx context.Context, data []byte) error {
	// TODO(mbm): make runner context aware
	// TODO(mbm): reduce duplication with cmd/benchrun

	// Parse job.
	j, err := job.Unmarshal(data)
	if err != nil {
		return err
	}

	// Build toolchain.
	tc, err := runner.NewToolchain(j.Toolchain.Type, j.Toolchain.Params)
	if err != nil {
		return err
	}
	lg.Param(h, "toolchain", tc.String())

	// GCS filesystem.
	store, err := gcs.New(ctx, bucket)
	if err != nil {
		return err
	}

	// Construct workspace.
	w, err := runner.NewWorkspace(
		runner.WithLogger(h),
		runner.WithArtifactStore(store),
	)
	if err != nil {
		return err
	}

	// Initialize runner.
	r := runner.NewRunner(w, tc)

	if err := h.ConfigureRunner(r); err != nil {
		return err
	}

	r.Init(ctx)

	// Run benchmarks.
	for _, s := range j.Suites {
		mod := runner.Module{
			Path:    s.Module.Path,
			Version: s.Module.Version,
		}
		r.Benchmark(ctx, runner.Job{Module: mod})
	}

	// Cleanup.
	r.Clean(ctx)

	return w.Error()
}
