package db

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/lib/pq"

	"github.com/mmcloughlin/cb/app/db/internal/db"
	"github.com/mmcloughlin/cb/app/entity"
)

// StoreCommit writes commit to the database.
func (d *DB) StoreCommit(ctx context.Context, c *entity.Commit) error {
	return d.txq(ctx, func(q *db.Queries) error {
		return storeCommit(ctx, q, c)
	})
}

func storeCommit(ctx context.Context, q *db.Queries, c *entity.Commit) error {
	sha, err := hex.DecodeString(c.SHA)
	if err != nil {
		return fmt.Errorf("invalid sha: %w", err)
	}

	tree, err := hex.DecodeString(c.Tree)
	if err != nil {
		return fmt.Errorf("invalid tree: %w", err)
	}

	parents := make([][]byte, len(c.Parents))
	for i, p := range c.Parents {
		parents[i], err = hex.DecodeString(p)
		if err != nil {
			return fmt.Errorf("invalid parent: %w", err)
		}
	}

	return q.InsertCommit(ctx, db.InsertCommitParams{
		SHA:            sha,
		Tree:           tree,
		Parents:        parents,
		AuthorName:     c.Author.Name,
		AuthorEmail:    c.Author.Email,
		AuthorTime:     c.AuthorTime,
		CommitterName:  c.Committer.Name,
		CommitterEmail: c.Committer.Email,
		CommitTime:     c.CommitTime,
		Message:        c.Message,
	})
}

// StoreCommits writes the given commits to the database in a single batch.
func (d *DB) StoreCommits(ctx context.Context, cs []*entity.Commit) error {
	fields := []string{
		"sha",
		"tree",
		"parents",
		"author_name",
		"author_email",
		"author_time",
		"committer_name",
		"committer_email",
		"commit_time",
		"message",
	}
	values := []interface{}{}
	for _, c := range cs {
		sha, err := hex.DecodeString(c.SHA)
		if err != nil {
			return fmt.Errorf("invalid sha: %w", err)
		}

		tree, err := hex.DecodeString(c.Tree)
		if err != nil {
			return fmt.Errorf("invalid tree: %w", err)
		}

		parents := make([][]byte, len(c.Parents))
		for i, p := range c.Parents {
			parents[i], err = hex.DecodeString(p)
			if err != nil {
				return fmt.Errorf("invalid parent: %w", err)
			}
		}

		values = append(values,
			sha,
			tree,
			pq.ByteaArray(parents),
			c.Author.Name,
			c.Author.Email,
			c.AuthorTime,
			c.Committer.Name,
			c.Committer.Email,
			c.CommitTime,
			c.Message,
		)
	}
	return d.tx(ctx, func(tx *sql.Tx) error {
		return d.insert(ctx, tx, "commits", fields, values)
	})
}

// FindCommitBySHA looks up the given commit in the database.
func (d *DB) FindCommitBySHA(ctx context.Context, sha string) (*entity.Commit, error) {
	shabytes, err := hex.DecodeString(sha)
	if err != nil {
		return nil, fmt.Errorf("invalid sha: %w", err)
	}

	var c *entity.Commit
	err = d.txq(ctx, func(q *db.Queries) error {
		var err error
		c, err = findCommitBySHA(ctx, q, shabytes)
		return err
	})
	return c, err
}

func findCommitBySHA(ctx context.Context, q *db.Queries, sha []byte) (*entity.Commit, error) {
	c, err := q.Commit(ctx, sha)
	if err != nil {
		return nil, err
	}

	return mapCommit(c), nil
}

// MostRecentCommit returns the most recent commit by commit time.
func (d *DB) MostRecentCommit(ctx context.Context) (*entity.Commit, error) {
	var c *entity.Commit
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		c, err = mostRecentCommit(ctx, q)
		return err
	})
	return c, err
}

func mostRecentCommit(ctx context.Context, q *db.Queries) (*entity.Commit, error) {
	c, err := q.MostRecentCommit(ctx)
	if err != nil {
		return nil, err
	}

	return mapCommit(c), nil
}

// MostRecentCommitWithRef returns the most recent commit by commit time having the supplied ref.
func (d *DB) MostRecentCommitWithRef(ctx context.Context, ref string) (*entity.Commit, error) {
	var c *entity.Commit
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		c, err = mostRecentCommitWithRef(ctx, q, ref)
		return err
	})
	return c, err
}

func mostRecentCommitWithRef(ctx context.Context, q *db.Queries, ref string) (*entity.Commit, error) {
	c, err := q.MostRecentCommitWithRef(ctx, ref)
	if err != nil {
		return nil, err
	}

	return mapCommit(c), nil
}

func mapCommit(c db.Commit) *entity.Commit {
	parents := make([]string, len(c.Parents))
	for i, parent := range c.Parents {
		parents[i] = hex.EncodeToString(parent)
	}

	return &entity.Commit{
		SHA:     hex.EncodeToString(c.SHA),
		Tree:    hex.EncodeToString(c.Tree),
		Parents: parents,
		Author: entity.Person{
			Name:  c.AuthorName,
			Email: c.AuthorEmail,
		},
		AuthorTime: c.AuthorTime,
		Committer: entity.Person{
			Name:  c.CommitterName,
			Email: c.CommitterEmail,
		},
		CommitTime: c.CommitTime,
		Message:    c.Message,
	}
}

// StoreCommitRef writes a commit ref pair to the database.
func (d *DB) StoreCommitRef(ctx context.Context, r *entity.CommitRef) error {
	return d.txq(ctx, func(q *db.Queries) error {
		return storeCommitRef(ctx, q, r)
	})
}

// StoreCommitRefs writes the given commit refs to the database.
func (d *DB) StoreCommitRefs(ctx context.Context, rs []*entity.CommitRef) error {
	fields := []string{
		"sha",
		"ref",
	}
	values := []interface{}{}
	for _, r := range rs {
		sha, err := hex.DecodeString(r.SHA)
		if err != nil {
			return fmt.Errorf("invalid sha: %w", err)
		}

		values = append(values,
			sha,
			r.Ref,
		)
	}
	return d.tx(ctx, func(tx *sql.Tx) error {
		return d.insert(ctx, tx, "commit_refs", fields, values)
	})
}

func storeCommitRef(ctx context.Context, q *db.Queries, r *entity.CommitRef) error {
	sha, err := hex.DecodeString(r.SHA)
	if err != nil {
		return fmt.Errorf("invalid sha: %w", err)
	}

	return q.InsertCommitRef(ctx, db.InsertCommitRefParams{
		SHA: sha,
		Ref: r.Ref,
	})
}

// StoreCommitPosition writes a commit position to the database. This should be
// rarely needed outside of testing; prefer BuildCommitPositions.
func (d *DB) StoreCommitPosition(ctx context.Context, p *entity.CommitPosition) error {
	return d.txq(ctx, func(q *db.Queries) error {
		return storeCommitPosition(ctx, q, p)
	})
}

func storeCommitPosition(ctx context.Context, q *db.Queries, p *entity.CommitPosition) error {
	sha, err := hex.DecodeString(p.SHA)
	if err != nil {
		return fmt.Errorf("invalid sha: %w", err)
	}

	return q.InsertCommitPosition(ctx, db.InsertCommitPositionParams{
		SHA:        sha,
		CommitTime: p.CommitTime,
		Index:      int32(p.Index),
	})
}

// BuildCommitPositions creates the commit positions table. The table is
// completely rebuilt from the source tables.
func (d *DB) BuildCommitPositions(ctx context.Context) error {
	return d.txq(ctx, func(q *db.Queries) error {
		return q.BuildCommitPositions(ctx)
	})
}

// MostRecentCommitIndex returns the most recent commit index.
func (d *DB) MostRecentCommitIndex(ctx context.Context) (int, error) {
	var idx int
	err := d.txq(ctx, func(q *db.Queries) error {
		i, err := q.MostRecentCommitIndex(ctx)
		idx = int(i)
		return err
	})
	return idx, err
}

// StoreModule writes module to the database.
func (d *DB) StoreModule(ctx context.Context, m *entity.Module) error {
	return d.txq(ctx, func(q *db.Queries) error {
		return storeModule(ctx, q, m)
	})
}
