package db

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/mmcloughlin/cb/app/db/internal/db"
	"github.com/mmcloughlin/cb/app/entity"
)

//go:generate sqlc generate

// DB provides a database storage layer.
type DB struct {
	db *sql.DB
	q  *db.Queries
}

// Open postgres database connection with the given connection string.
func Open(conn string) (*DB, error) {
	d, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	return &DB{
		db: d,
		q:  db.New(d),
	}, nil
}

// Close database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// StoreCommit writes commit to the database.
func (d *DB) StoreCommit(ctx context.Context, c *entity.Commit) error {
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

	return d.q.InsertCommit(ctx, db.InsertCommitParams{
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

// FindCommitBySHA looks up the given commit in the database.
func (d *DB) FindCommitBySHA(ctx context.Context, sha string) (*entity.Commit, error) {
	shabytes, err := hex.DecodeString(sha)
	if err != nil {
		return nil, fmt.Errorf("invalid sha: %w", err)
	}

	c, err := d.q.Commit(ctx, shabytes)
	if err != nil {
		return nil, err
	}

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
	}, nil
}

// StoreModule writes module to the database.
func (d *DB) StoreModule(ctx context.Context, m *entity.Module) error {
	return d.q.InsertModule(ctx, db.InsertModuleParams{
		UUID:    m.UUID(),
		Path:    m.Path,
		Version: m.Version,
	})
}

// FindModuleByUUID looks up the given module in the database.
func (d *DB) FindModuleByUUID(ctx context.Context, id uuid.UUID) (*entity.Module, error) {
	m, err := d.q.Module(ctx, id)
	if err != nil {
		return nil, err
	}

	return &entity.Module{
		Path:    m.Path,
		Version: m.Version,
	}, nil
}
