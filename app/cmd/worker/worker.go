package main

import (
	"context"
	"log"
	"os"

	"github.com/mmcloughlin/cb/app/consumer"
	"github.com/mmcloughlin/cb/app/gcs"
	"github.com/mmcloughlin/cb/app/job"
	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/lg"
	"github.com/mmcloughlin/cb/pkg/runner"
)

var (
	// TODO(mbm): remove hardcoded subscription
	subscription = "projects/contbench/subscriptions/worker_jobs"
	// TODO(mbm): remove hardcoded bucket
	bucket = "contbench_results"
)

func main() {
	os.Exit(main1())
}

func main1() int {
	if err := mainerr(); err != nil {
		log.Print(err)
		return 1
	}
	return 0
}

func mainerr() error {
	l := lg.Default()
	ctx := command.BackgroundContext(l)
	h := &Handler{Logger: l}
	c, err := consumer.New(ctx, subscription, h, consumer.WithLogger(l))
	if err != nil {
		return err
	}
	defer c.Close()
	return c.Receive(ctx)
}

type Handler struct {
	lg.Logger
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
