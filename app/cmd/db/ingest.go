package main

import (
	"context"
	"flag"

	"github.com/google/subcommands"

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

	d := db.New(sqldb)
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
		cmd.Log.Printf("file=%s mod=%s", file.Path, file.ModTime)

		rs, err := loader.Load(ctx, file.Path)
		if err != nil {
			cmd.Log.Printf("loading error: %s", err)
			continue
		}
		cmd.Log.Printf("loaded %d results", len(rs))

		for _, r := range rs {
			if err := d.StoreResult(ctx, r); err != nil {
				return cmd.Error(err)
			}
		}
		cmd.Log.Printf("inserted %d results", len(rs))
	}

	return subcommands.ExitSuccess
}
