package main

import (
	"context"
	"flag"
	"io"
	"net/http"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/app/coordinator"
	"github.com/mmcloughlin/cb/app/worker"
	"github.com/mmcloughlin/cb/internal/errutil"
	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/platform"
)

type Run struct {
	command.Base
	*platform.Platform

	name           string
	coordinatorURL string
}

func NewRun(b command.Base, p *platform.Platform) *Run {
	return &Run{
		Base:     b,
		Platform: p,
	}
}

func (*Run) Name() string { return "run" }

func (*Run) Synopsis() string {
	return "run benchmark worker"
}

func (*Run) Usage() string {
	return ""
}

func (cmd *Run) SetFlags(f *flag.FlagSet) {
	cmd.Platform.SetFlags(f)

	f.StringVar(&cmd.name, "name", "", "worker name")
	f.StringVar(&cmd.coordinatorURL, "coordinator", "", "coordinator address")
}

func (cmd *Run) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	c := coordinator.NewClient(http.DefaultClient, cmd.coordinatorURL, cmd.name)
	p := &Processor{}
	w := worker.New(c, p, worker.WithLogger(cmd.Log))
	return cmd.Status(w.Run(ctx))
}

type Processor struct{}

func (p *Processor) Process(ctx context.Context, j *coordinator.Job) (io.ReadCloser, error) {
	return nil, errutil.ErrNotImplemented
}
