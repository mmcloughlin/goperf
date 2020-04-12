package db

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/mmcloughlin/cb/app/db/internal/db"
	"github.com/mmcloughlin/cb/app/entity"
)

// ListTaskSpecsRecentCommitsWithoutWorkerResults searches for recent commits without results for a given worker.
func (d *DB) ListTaskSpecsRecentCommitsWithoutWorkerResults(ctx context.Context, worker string, since time.Time, n int) ([]entity.TaskSpec, error) {
	var specs []entity.TaskSpec
	err := d.tx(ctx, func(q *db.Queries) error {
		var err error
		specs, err = listTaskSpecsRecentCommitsWithoutWorkerResults(ctx, q, worker, since, n)
		return err
	})
	return specs, err
}

func listTaskSpecsRecentCommitsWithoutWorkerResults(ctx context.Context, q *db.Queries, worker string, since time.Time, n int) ([]entity.TaskSpec, error) {
	rows, err := q.RecentCommitModulePairsWithoutWorkerResults(ctx, db.RecentCommitModulePairsWithoutWorkerResultsParams{
		Worker: worker,
		Since:  since,
		Num:    int32(n),
	})
	if err != nil {
		return nil, err
	}

	specs := make([]entity.TaskSpec, len(rows))
	for i, row := range rows {
		specs[i] = entity.TaskSpec{
			CommitSHA:  hex.EncodeToString(row.CommitSHA),
			Type:       entity.TaskTypeModule,
			TargetUUID: row.ModuleUUID,
		}
	}

	return specs, nil
}
