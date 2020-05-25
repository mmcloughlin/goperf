package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"

	"github.com/google/subcommands"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/mmcloughlin/goperf/app/change/changetest"
	"github.com/mmcloughlin/goperf/app/db"
	"github.com/mmcloughlin/goperf/app/entity"
	"github.com/mmcloughlin/goperf/app/trace"
	"github.com/mmcloughlin/goperf/pkg/command"
)

type ChangeTest struct {
	command.Base

	benchmarkUUID   string
	environmentUUID string
	index           int
	context         int
	output          string
}

func NewChangeTest(b command.Base) *ChangeTest {
	return &ChangeTest{
		Base: b,
	}
}

func (*ChangeTest) Name() string { return "changetest" }

func (*ChangeTest) Synopsis() string {
	return "initialize change detection test case"
}

func (*ChangeTest) Usage() string {
	return ""
}

func (cmd *ChangeTest) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.benchmarkUUID, "benchmark-uuid", "", "benchmark uuid")
	f.StringVar(&cmd.environmentUUID, "environment-uuid", "", "environment uuid")
	f.IntVar(&cmd.index, "commit-index", -1, "commit index")
	f.IntVar(&cmd.context, "context", 150, "number of commits either side")
	f.StringVar(&cmd.output, "output", ".", "output directory")
}

func (cmd *ChangeTest) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) (status subcommands.ExitStatus) {
	// Parse options.
	benchmarkUUID, err := uuid.Parse(cmd.benchmarkUUID)
	if err != nil {
		return cmd.UsageError("benchmark uuid: %v", err)
	}

	environmentUUID, err := uuid.Parse(cmd.environmentUUID)
	if err != nil {
		return cmd.UsageError("environment uuid: %v", err)
	}

	if cmd.index < 0 {
		return cmd.UsageError("must provide commit index")
	}

	if info, err := os.Stat(cmd.output); err != nil || !info.IsDir() {
		return cmd.UsageError("must specify output directory")
	}

	id := trace.ID{
		BenchmarkUUID:   benchmarkUUID,
		EnvironmentUUID: environmentUUID,
	}

	r := entity.CommitIndexRange{
		Min: cmd.index - cmd.context,
		Max: cmd.index + cmd.context,
	}

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

	// Query for trace.
	cmd.Log.Info("fetching trace",
		zap.Stringer("id", id),
		zap.Stringer("commits", r),
	)

	t, err := d.Trace(ctx, id, r)
	if err != nil {
		return cmd.Error(err)
	}

	// Write to file.
	filename := filepath.Join(cmd.output, changetest.Filename(id))
	cmd.Log.Info("write file", zap.String("output", filename))
	if err := changetest.WriteNewCaseFile(filename, t.Series); err != nil {
		return cmd.Error(err)
	}

	return subcommands.ExitSuccess
}
