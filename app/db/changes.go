package db

import (
	"context"
	"database/sql"

	"github.com/mmcloughlin/cb/app/entity"
)

// StoreChangesBatch writes the given changes to the database in a single batch.
// Does not write any dependent objects.
func (d *DB) StoreChangesBatch(ctx context.Context, cs []*entity.Change) error {
	fields := []string{
		"benchmark_uuid",
		"index",
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
			c.Benchmark.UUID(),
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
	return d.tx(ctx, func(tx *sql.Tx) error {
		return d.insert(ctx, tx, "changes", fields, values, "ON CONFLICT DO NOTHING")
	})
}
