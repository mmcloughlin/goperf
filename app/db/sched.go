package db

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/app/db/internal/db"
	"github.com/mmcloughlin/cb/app/entity"
)

// CommitModule represents a commit module pair.
type CommitModule struct {
	CommitSHA  string
	CommitTime time.Time
	ModuleUUID uuid.UUID
}

// ListCommitModulesWithoutCompleteTasks searches for n recent commit module
// pairs without completed tasks for the given worker.
func (d *DB) ListCommitModulesWithoutCompleteTasks(ctx context.Context, worker string, n int) ([]CommitModule, error) {
	var cms []CommitModule
	err := d.tx(ctx, func(q *db.Queries) error {
		var err error
		cms, err = listCommitModulesWithoutTasksInStatus(ctx, q, worker, entity.TaskStatusCompleteValues(), n)
		return err
	})
	return cms, err
}

func listCommitModulesWithoutTasksInStatus(ctx context.Context, q *db.Queries, worker string, statuses []entity.TaskStatus, n int) ([]CommitModule, error) {
	s, err := toTaskStatuses(statuses)
	if err != nil {
		return nil, err
	}

	rows, err := q.RecentCommitModulePairsWithoutWorkerTasks(ctx, db.RecentCommitModulePairsWithoutWorkerTasksParams{
		Worker:   worker,
		Statuses: s,
		Num:      int32(n),
	})
	if err != nil {
		return nil, err
	}

	cms := make([]CommitModule, len(rows))
	for i, row := range rows {
		cms[i] = CommitModule{
			CommitSHA:  hex.EncodeToString(row.CommitSHA),
			CommitTime: row.CommitTime,
			ModuleUUID: row.ModuleUUID,
		}
	}

	return cms, nil
}
