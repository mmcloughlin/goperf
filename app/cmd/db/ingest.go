package main

import (
	"context"
	"flag"

	"github.com/google/subcommands"
	"go.uber.org/zap"

	"github.com/mmcloughlin/cb/app/db"
	"github.com/mmcloughlin/cb/app/gcs"
	"github.com/mmcloughlin/cb/app/results"
	"github.com/mmcloughlin/cb/pkg/command"
)

type Ingest struct {
	command.Base

	bucket string
}

func NewIngest(b command.Base) *Ingest {
	return &Ingest{
		Base: b,
	}
}

func (*Ingest) Name() string { return "ingest" }

func (*Ingest) Synopsis() string {
	return "ingest result files"
}

func (*Ingest) Usage() string {
	return ""
}

func (cmd *Ingest) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.bucket, "bucket", "", "data files bucket")
}

func (cmd *Ingest) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	// Open database.
	sqldb, err := open()
	if err != nil {
		return cmd.Error(err)
	}

	d, err := db.New(ctx, sqldb)
	if err != nil {
		return cmd.Error(err)
	}
	defer d.Close()

	// Open filesystem.
	fs, err := gcs.New(ctx, cmd.bucket)
	if err != nil {
		return cmd.Error(err)
	}

	// List files.
	files, err := fs.List(ctx, "")
	if err != nil {
		return cmd.Error(err)
	}

	// Extract results.
	loader, err := results.NewLoader(results.WithFilesystem(fs))
	if err != nil {
		return cmd.Error(err)
	}

	for _, file := range files {
		cmd.Log.Info("ingest file", zap.String("file", file.Path), zap.Time("modtime", file.ModTime))

		rs, err := loader.Load(ctx, file.Path)
		if err != nil {
			cmd.Log.Error("loading error", zap.Error(err))
			continue
		}
		cmd.Log.Info("loaded results", zap.Int("num_results", len(rs)))

		for _, r := range rs {
			if err := d.StoreResult(ctx, r); err != nil {
				return cmd.Error(err)
			}
		}
		cmd.Log.Info("inserted results", zap.Int("num_results", len(rs)))
	}

	return subcommands.ExitSuccess
}
