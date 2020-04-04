package main

import (
	"context"
	"flag"
	"io"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/app/db"
	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/app/repo"
	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/lg"
)

type Commits struct {
	command.Base

	batch int
}

func NewCommits(b command.Base) *Commits {
	return &Commits{
		Base: b,
	}
}

func (*Commits) Name() string { return "commits" }

func (*Commits) Synopsis() string {
	return "import all go repository commits"
}

func (*Commits) Usage() string {
	return ""
}

func (cmd *Commits) SetFlags(f *flag.FlagSet) {
	f.IntVar(&cmd.batch, "batch", 1024, "number of inserts per transaction")
}

func (cmd *Commits) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
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

	// Clone the Go repository.
	scope := lg.Scope(cmd.Log, "clone")
	g, err := repo.Clone(ctx, repo.GoURL)
	scope()
	if err != nil {
		return cmd.Error(err)
	}
	defer g.Close()

	// "git log"
	scope = lg.Scope(cmd.Log, "git log")
	it, err := g.Log()
	scope()
	if err != nil {
		return cmd.Error(err)
	}
	defer it.Close()

	n := 0
	batch := []*entity.Commit{}
	for {
		c, err := it.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cmd.Error(err)
		}

		// Insert if the batch is full.
		if len(batch) == cmd.batch {
			cmd.Log.Printf("inserting %d commits", len(batch))
			if err := d.StoreCommits(ctx, batch); err != nil {
				return cmd.Error(err)
			}
			batch = batch[:0]
		}

		// Add to batch.
		batch = append(batch, c)
		n++
		cmd.Log.Printf("added %s total %d", c.SHA, n)
	}

	// Final batch.
	cmd.Log.Printf("inserting %d commits", len(batch))
	if err := d.StoreCommits(ctx, batch); err != nil {
		return cmd.Error(err)
	}

	return subcommands.ExitSuccess
}
