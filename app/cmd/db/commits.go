package main

import (
	"context"
	"flag"
	"io"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/app/db"
	"github.com/mmcloughlin/cb/app/repo"
	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/lg"
)

type Commits struct {
	command.Base
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
}

func (cmd *Commits) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	// Open database.
	sqldb, err := open()
	if err != nil {
		return cmd.Error(err)
	}

	d := db.New(sqldb)
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
	for {
		c, err := it.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cmd.Error(err)
		}

		// Insert.
		if err := d.StoreCommit(ctx, c); err != nil {
			return cmd.Error(err)
		}
		n++

		cmd.Log.Printf("inserted %s total %d", c.SHA, n)
	}

	return subcommands.ExitSuccess
}
