package db

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/mmcloughlin/cb/app/change"
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

// ListChangeSummaries returns changes with associated metadata.
func (d *DB) ListChangeSummaries(ctx context.Context, r entity.CommitIndexRange, minEffectSize float64) ([]*entity.ChangeSummary, error) {
	var cs []*entity.ChangeSummary
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		cs, err = listChangeSummaries(ctx, q, r, minEffectSize)
		return err
	})
	return cs, err
}

func listChangeSummaries(ctx context.Context, q *db.Queries, r entity.CommitIndexRange, minEffectSize float64) ([]*entity.ChangeSummary, error) {
	rows, err := q.ChangeSummaries(ctx, db.ChangeSummariesParams{
		EffectSizeMin:  minEffectSize,
		CommitIndexMin: int32(r.Min),
		CommitIndexMax: int32(r.Max),
	})
	if err != nil {
		return nil, err
	}

	cs := make([]*entity.ChangeSummary, len(rows))
	for i, row := range rows {
		params := map[string]string{}
		if err := json.Unmarshal(row.Parameters, &params); err != nil {
			return nil, fmt.Errorf("decode parameters: %w", err)
		}

		cs[i] = &entity.ChangeSummary{
			Benchmark: &entity.Benchmark{
				Package: &entity.Package{
					Module: &entity.Module{
						Path:    row.Path,
						Version: row.Version,
					},
					RelativePath: row.RelativePath,
				},
				FullName:   row.FullName,
				Name:       row.Name,
				Parameters: params,
				Unit:       row.Unit,
			},
			EnvironmentUUID: row.EnvironmentUUID,
			CommitSHA:       hex.EncodeToString(row.CommitSHA),
			CommitSubject:   row.CommitSubject,
			Change: change.Change{
				CommitIndex: int(row.CommitIndex),
				EffectSize:  row.EffectSize,
				Pre: change.Stats{
					N:        int(row.PreN),
					Mean:     row.PreMean,
					Variance: row.PreStddev * row.PreStddev,
				},
				Post: change.Stats{
					N:        int(row.PostN),
					Mean:     row.PostMean,
					Variance: row.PostStddev * row.PostStddev,
				},
			},
		}
	}

	return cs, nil
}
