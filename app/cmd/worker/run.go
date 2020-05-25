package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"net/http"
	"os"

	"github.com/google/subcommands"
	"go.uber.org/zap"

	"github.com/mmcloughlin/goperf/app/coordinator"
	"github.com/mmcloughlin/goperf/app/worker"
	"github.com/mmcloughlin/goperf/pkg/command"
	"github.com/mmcloughlin/goperf/pkg/fs"
	"github.com/mmcloughlin/goperf/pkg/platform"
	"github.com/mmcloughlin/goperf/pkg/runner"
)

type Run struct {
	command.Base
	*platform.Platform

	name           string
	coordinatorURL string
	artifacts      string
	goproxy        string
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

	defaultName := ""
	if hostname, err := os.Hostname(); err == nil {
		defaultName = hostname
	}

	f.StringVar(&cmd.name, "name", defaultName, "worker name")
	f.StringVar(&cmd.coordinatorURL, "coordinator", "", "coordinator address")
	f.StringVar(&cmd.artifacts, "artifacts", "", "artifacts storage directory")
	f.StringVar(&cmd.goproxy, "goproxy", "proxy.golang.org", "GOPROXY environment variable")
}

func (cmd *Run) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	c := coordinator.NewClient(http.DefaultClient, cmd.coordinatorURL, cmd.name)

	artifacts := fs.NewLocal(cmd.artifacts)

	p := &Processor{
		platform:  cmd.Platform,
		artifacts: artifacts,
		goproxy:   cmd.goproxy,
		log:       cmd.Log,
	}

	w := worker.New(c, p, worker.WithLogger(cmd.Log))

	return cmd.Status(w.Run(ctx))
}

type Processor struct {
	platform  *platform.Platform
	artifacts fs.Interface
	goproxy   string
	log       *zap.Logger
}

func (p *Processor) Process(ctx context.Context, j *coordinator.Job) (io.ReadCloser, error) {
	// TODO(mbm): make runner context aware
	// TODO(mbm): reduce duplication with cmd/benchrun

	// Build toolchain.
	builderType, ok := runner.HostSnapshotBuilderType()
	if !ok {
		return nil, errors.New("could not identify builder type")
	}
	tc := runner.NewSnapshot(builderType, j.CommitSHA)

	p.log.Info("constructed toolchain", zap.Stringer("toolchain", tc))

	// Construct workspace.
	w, err := runner.NewWorkspace(
		runner.WithLogger(p.log),
		runner.WithArtifactStore(p.artifacts),
	)
	if err != nil {
		return nil, err
	}

	// Initialize runner.
	r := runner.NewRunner(w, tc)
	r.SetGoProxy(p.goproxy)
	if err := p.platform.ConfigureRunner(r); err != nil {
		return nil, err
	}

	r.Init(ctx)

	// Run benchmark.
	output := j.UUID.String()
	r.Benchmark(ctx, j.Suite, output)

	// Cleanup.
	r.Clean(ctx)

	// Bail if there was some error.
	if err := w.Error(); err != nil {
		return nil, err
	}

	// Otherwise return a handle to the output.
	return p.artifacts.Open(ctx, output)
}
