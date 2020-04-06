// Code generated by sqlc. DO NOT EDIT.
// source: tasks.sql

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const workerTasksWithSpecAndStatus = `-- name: WorkerTasksWithSpecAndStatus :many
SELECT
    uuid, worker, commit_sha, type, target_uuid, status, last_status_update, datafile_uuid
FROM
    tasks
WHERE 1=1
    AND worker=$1
    AND type=$2
    AND target_uuid=$3
    AND commit_sha=$4
    AND status = ANY ($5::task_status[])
`

type WorkerTasksWithSpecAndStatusParams struct {
	Worker     string
	Type       TaskType
	TargetUUID uuid.UUID
	CommitSHA  []byte
	Statuses   []TaskStatus
}

func (q *Queries) WorkerTasksWithSpecAndStatus(ctx context.Context, arg WorkerTasksWithSpecAndStatusParams) ([]Task, error) {
	rows, err := q.query(ctx, q.workerTasksWithSpecAndStatusStmt, workerTasksWithSpecAndStatus,
		arg.Worker,
		arg.Type,
		arg.TargetUUID,
		arg.CommitSHA,
		pq.Array(arg.Statuses),
	)
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
