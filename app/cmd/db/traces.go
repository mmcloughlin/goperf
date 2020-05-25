package main

import (
	"context"
	"flag"

	"github.com/google/subcommands"
	"go.uber.org/zap"

	"github.com/mmcloughlin/goperf/app/db"
	"github.com/mmcloughlin/goperf/app/entity"
	"github.com/mmcloughlin/goperf/app/trace"
	"github.com/mmcloughlin/goperf/pkg/command"
)

type Traces struct {
	command.Base

	output string
	num    int
}

func NewTraces(b command.Base) *Traces {
	return &Traces{
		Base: b,
	}
}

func (*Traces) Name() string { return "traces" }

func (*Traces) Synopsis() string {
	return "download traces"
}

func (*Traces) Usage() string {
	return ""
}

func (cmd *Traces) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.output, "output", "traces.csv.gz", "output file")
	f.IntVar(&cmd.num, "num", 150, "number of most recent commits")
}

func (cmd *Traces) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) (status subcommands.ExitStatus) {
	// Open database.
	sqldb, err := open()
	if err != nil {
		return cmd.Error(err)
	}

	d, err := db.New(ctx, sqldb)
	if err != nil {
		return cmd.Error(err)
	}
	defer cmd.CheckClose(&status, d)

	// Determine commit index.
	idx, err := d.MostRecentCommitIndex(ctx)
	if err != nil {
		return cmd.Error(err)
	}
	cmd.Log.Info("most recent commit index", zap.Int("index", idx))

	r := entity.CommitIndexRange{
		Min: idx - cmd.num + 1,
		Max: idx,
	}

	// Query for trace points.
	cmd.Log.Info("fetching traces",
		zap.Int("min_commit_index", r.Min),
		zap.Int("max_commit_index", r.Max),
	)

	ps, err := d.ListTracePoints(ctx, r)
	if err != nil {
		return cmd.Error(err)
	}

	// Write to file.
	cmd.Log.Info("write file", zap.String("output", cmd.output))
	if err := trace.WritePointsFile(cmd.output, ps); err != nil {
		return cmd.Error(err)
	}

	return subcommands.ExitSuccess
}
