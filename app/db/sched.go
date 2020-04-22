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
	err := d.txq(ctx, func(q *db.Queries) error {
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

// CommitModuleError represents a commit module pair that has no completed tasks
// and at least one error.
type CommitModuleError struct {
	ModuleUUID  uuid.UUID
	CommitSHA   string
	NumErrors   int
	LastAttempt time.Time
}

// ListCommitModuleErrors returns up to n commit module pairs that have errored
// on the given worker with no successful execution. The search is limited to
// pairs with at most maxErrors errors and last attempt before the given
// timestamp. This is intended for identifying tasks that should be retried.
func (d *DB) ListCommitModuleErrors(ctx context.Context, worker string, maxErrors int, lastAttempt time.Time, n int) ([]CommitModuleError, error) {
	var results []CommitModuleError
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		results, err = listCommitModuleErrors(ctx, q, worker, maxErrors, lastAttempt, n)
		return err
	})
	return results, err
}

func listCommitModuleErrors(ctx context.Context, q *db.Queries, worker string, maxErrors int, lastAttempt time.Time, n int) ([]CommitModuleError, error) {
	rows, err := q.CommitModuleWorkerErrors(ctx, db.CommitModuleWorkerErrorsParams{
		Worker:            worker,
		MaxErrors:         int32(maxErrors),
		LastAttemptBefore: lastAttempt,
		Num:               int32(n),
	})
	if err != nil {
		return nil, err
	}

	results := make([]CommitModuleError, len(rows))
	for i, row := range rows {
		results[i] = CommitModuleError{
			ModuleUUID:  row.ModuleUUID,
			CommitSHA:   hex.EncodeToString(row.CommitSHA),
			NumErrors:   int(row.NumErrors),
			LastAttempt: row.LastAttemptTime,
		}
	}

	return results, nil
}
