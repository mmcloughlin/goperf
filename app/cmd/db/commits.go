package main

import (
	"context"
	"flag"
	"io"

	"github.com/google/subcommands"
	"go.uber.org/zap"

	"github.com/mmcloughlin/goperf/app/db"
	"github.com/mmcloughlin/goperf/app/entity"
	"github.com/mmcloughlin/goperf/app/repo"
	"github.com/mmcloughlin/goperf/pkg/command"
	"github.com/mmcloughlin/goperf/pkg/lg"
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
	f.IntVar(&cmd.batch, "batch", 1024, "number of inserts per batch")
}

func (cmd *Commits) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) (status subcommands.ExitStatus) {
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

	// Clone the Go repository.
	scope := lg.Scope(cmd.Log, "clone")
	g, err := repo.Clone(ctx, repo.GoURL)
	scope()
	if err != nil {
		return cmd.Error(err)
	}
	defer cmd.CheckClose(&status, g)

	// "git log"
	scope = lg.Scope(cmd.Log, "git log")
	it, err := g.Log()
	scope()
	if err != nil {
		return cmd.Error(err)
	}
	defer it.Close()

	batches := NewBatchIterator(it, cmd.batch)
	for {
		batch, err := batches.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cmd.Error(err)
		}

		// Insert batch.
		cmd.Log.Info("inserting commits", zap.Int("num_commits", len(batch)))
		if err := d.StoreCommits(ctx, batch); err != nil {
			return cmd.Error(err)
		}
	}

	return subcommands.ExitSuccess
}

type Refs struct {
	command.Base

	batch int
}

func NewRefs(b command.Base) *Refs {
	return &Refs{
		Base: b,
	}
}

func (*Refs) Name() string { return "refs" }

func (*Refs) Synopsis() string {
	return "insert commit ref mappings"
}

func (*Refs) Usage() string {
	return ""
}

func (cmd *Refs) SetFlags(f *flag.FlagSet) {
	f.IntVar(&cmd.batch, "batch", 1024, "number of inserts per batch")
}

func (cmd *Refs) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) (status subcommands.ExitStatus) {
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

	// Clone the Go repository.
	scope := lg.Scope(cmd.Log, "clone")
	g, err := repo.Clone(ctx, repo.GoURL)
	scope()
	if err != nil {
		return cmd.Error(err)
	}
	defer cmd.CheckClose(&status, g)

	// "git log --first-parent"
	scope = lg.Scope(cmd.Log, "git log")
	it, err := g.Log()
	scope()
	if err != nil {
		return cmd.Error(err)
	}

	it = repo.FirstParent(it)
	defer it.Close()

	ref := "master"
	batches := NewBatchIterator(it, cmd.batch)
	for {
		batch, err := batches.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cmd.Error(err)
		}

		// Convert to refs and insert.
		refs := make([]*entity.CommitRef, len(batch))
		for i, c := range batch {
			refs[i] = &entity.CommitRef{
				SHA: c.SHA,
				Ref: ref,
			}
		}

		cmd.Log.Info("inserting commit refs", zap.Int("num_commit_refs", len(refs)))
		if err := d.StoreCommitRefs(ctx, refs); err != nil {
			return cmd.Error(err)
		}
	}

	return subcommands.ExitSuccess
}

// BatchIterator is a helper for batching a commit sequence.
type BatchIterator struct {
	commits repo.CommitIterator
	n       int
}

// NewBatchIterator builds an iterator over batches of size n from commits.
func NewBatchIterator(commits repo.CommitIterator, n int) *BatchIterator {
	return &BatchIterator{
		commits: commits,
		n:       n,
	}
}

// Next returns the next batch.
func (b *BatchIterator) Next() ([]*entity.Commit, error) {
	batch := make([]*entity.Commit, 0, b.n)
	for len(batch) < b.n {
		c, err := b.commits.Next()
		switch {
		case err == io.EOF && len(batch) == 0:
			return nil, io.EOF
		case err == io.EOF && len(batch) > 0:
			return batch, nil
		case err != nil:
			return nil, err
		default:
			batch = append(batch, c)
		}
	}
	return batch, nil
}

type Positions struct {
	command.Base
}

func NewPositions(b command.Base) *Positions {
	return &Positions{
		Base: b,
	}
}

func (*Positions) Name() string { return "positions" }

func (*Positions) Synopsis() string {
	return "rebuild commit positions table"
}

func (*Positions) Usage() string {
	return ""
}

func (cmd *Positions) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) (status subcommands.ExitStatus) {
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

	// Rebuild.
	if err := d.BuildCommitPositions(ctx); err != nil {
		return cmd.Error(err)
	}

	return subcommands.ExitSuccess
}
