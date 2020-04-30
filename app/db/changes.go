package db

import (
	"context"
	"database/sql"

	"github.com/mmcloughlin/cb/app/db/internal/db"
	"github.com/mmcloughlin/cb/app/entity"
)

// StoreChangesBatch writes the given changes to the database in a single batch.
// Does not write any dependent objects.
func (d *DB) StoreChangesBatch(ctx context.Context, cs []*entity.Change) error {
	return d.tx(ctx, func(tx *sql.Tx) error {
		return d.storeChangesBatch(ctx, tx, cs)
	})
}

// ReplaceChanges transactionally deletes changes in a range and inserts supplied changes.
func (d *DB) ReplaceChanges(ctx context.Context, r entity.CommitIndexRange, cs []*entity.Change) error {
	return d.tx(ctx, func(tx *sql.Tx) error {
		if err := d.q.WithTx(tx).DeleteChangesCommitRange(ctx, db.DeleteChangesCommitRangeParams{
			CommitIndexMin: int32(r.Min),
			CommitIndexMax: int32(r.Max),
		}); err != nil {
			return err
		}

		return d.storeChangesBatch(ctx, tx, cs)
	})
}

func (d *DB) storeChangesBatch(ctx context.Context, tx *sql.Tx, cs []*entity.Change) error {
	fields := []string{
		"benchmark_uuid",
		"environment_uuid",
		"commit_index",
		"effect_size",
		"pre_n",
		"pre_mean",
		"pre_stddev",
		"post_n",
		"post_mean",
		"post_stddev",
	}
	values := []interface{}{}
	for _, c := range cs {
		values = append(values,
			c.BenchmarkUUID,
			c.EnvironmentUUID,
			c.CommitIndex,
			c.EffectSize,
			c.Pre.N,
			c.Pre.Mean,
			c.Pre.Stddev(),
			c.Post.N,
			c.Post.Mean,
			c.Post.Stddev(),
		)
	}
	return d.insert(ctx, tx, "changes", fields, values, "ON CONFLICT DO NOTHING")
}
