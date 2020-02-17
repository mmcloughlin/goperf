package main

import (
	"context"
	"log"
	"os"

	"github.com/mmcloughlin/cb/app/consumer"
	"github.com/mmcloughlin/cb/app/job"
	"github.com/mmcloughlin/cb/pkg/lg"
	"github.com/mmcloughlin/cb/pkg/runner"
)

var (
	// TODO(mbm): remove hardcoded subscription
	subscription = "projects/contbench/subscriptions/worker_jobs"
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
	ctx := context.Background()
	l := lg.Default()
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

func (h *Handler) Handle(_ context.Context, data []byte) error {
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

	// Construct workspace.
	// TODO(mbm): get it working without inheriting environment
	w, err := runner.NewWorkspace(runner.InheritEnviron(), runner.WithLogger(h))
	if err != nil {
		return err
	}

	// Initialize runner.
	r := runner.NewRunner(w, tc)
	if err := r.Init(); err != nil {
		return err
	}

	// Run benchmarks.
	for _, s := range j.Suites {
		mod := runner.Module{
			Path:    s.Module.Path,
			Version: s.Module.Version,
		}
		if err := r.Benchmark(runner.Job{Module: mod}); err != nil {
			return err
		}
	}

	return nil
}
