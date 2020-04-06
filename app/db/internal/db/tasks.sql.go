// Code generated by sqlc. DO NOT EDIT.
// source: tasks.sql

package db

import (
	"context"

	"github.com/lib/pq"
)

const workerTasksWithStatus = `-- name: WorkerTasksWithStatus :many
SELECT
    uuid, worker, commit_sha, type, target_uuid, status, last_status_update, datafile_uuid
FROM
    tasks
WHERE 1=1
    AND worker=$1
    AND status = ANY ($2::task_status[])
`

type WorkerTasksWithStatusParams struct {
	Worker   string
	Statuses []TaskStatus
}

func (q *Queries) WorkerTasksWithStatus(ctx context.Context, arg WorkerTasksWithStatusParams) ([]Task, error) {
	rows, err := q.query(ctx, q.workerTasksWithStatusStmt, workerTasksWithStatus, arg.Worker, pq.Array(arg.Statuses))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Task
	for rows.Next() {
		var i Task
		if err := rows.Scan(
			&i.UUID,
			&i.Worker,
			&i.CommitSHA,
			&i.Type,
			&i.TargetUUID,
			&i.Status,
			&i.LastStatusUpdate,
			&i.DatafileUUID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
