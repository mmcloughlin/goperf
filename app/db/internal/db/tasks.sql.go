// Code generated by sqlc. DO NOT EDIT.
// source: tasks.sql

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const createTask = `-- name: CreateTask :one
INSERT INTO tasks (
    uuid,
    worker,
    commit_sha,
    type,
    target_uuid,
    status,
    last_status_update
)VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    'created',
    NOW()
)
RETURNING uuid, worker, commit_sha, type, target_uuid, status, last_status_update, datafile_uuid
`

type CreateTaskParams struct {
	UUID       uuid.UUID
	Worker     string
	CommitSHA  []byte
	Type       TaskType
	TargetUUID uuid.UUID
}

func (q *Queries) CreateTask(ctx context.Context, arg CreateTaskParams) (Task, error) {
	row := q.queryRow(ctx, q.createTaskStmt, createTask,
		arg.UUID,
		arg.Worker,
		arg.CommitSHA,
		arg.Type,
		arg.TargetUUID,
	)
	var i Task
	err := row.Scan(
		&i.UUID,
		&i.Worker,
		&i.CommitSHA,
		&i.Type,
		&i.TargetUUID,
		&i.Status,
		&i.LastStatusUpdate,
		&i.DatafileUUID,
	)
	return i, err
}

const setTaskDataFile = `-- name: SetTaskDataFile :exec
UPDATE
    tasks
SET
    datafile_uuid = $1
WHERE
    uuid = $2
`

type SetTaskDataFileParams struct {
	DatafileUUID uuid.UUID
	UUID         uuid.UUID
}

func (q *Queries) SetTaskDataFile(ctx context.Context, arg SetTaskDataFileParams) error {
	_, err := q.exec(ctx, q.setTaskDataFileStmt, setTaskDataFile, arg.DatafileUUID, arg.UUID)
	return err
}

const task = `-- name: Task :one
SELECT uuid, worker, commit_sha, type, target_uuid, status, last_status_update, datafile_uuid FROM tasks
WHERE uuid = $1 LIMIT 1
`

func (q *Queries) Task(ctx context.Context, uuid uuid.UUID) (Task, error) {
	row := q.queryRow(ctx, q.taskStmt, task, uuid)
	var i Task
	err := row.Scan(
		&i.UUID,
		&i.Worker,
		&i.CommitSHA,
		&i.Type,
		&i.TargetUUID,
		&i.Status,
		&i.LastStatusUpdate,
		&i.DatafileUUID,
	)
	return i, err
}

const tasksWithStatus = `-- name: TasksWithStatus :many
SELECT
    uuid, worker, commit_sha, type, target_uuid, status, last_status_update, datafile_uuid
FROM
    tasks
WHERE 1=1
    AND status = ANY ($1::task_status[])
`

func (q *Queries) TasksWithStatus(ctx context.Context, statuses []TaskStatus) ([]Task, error) {
	rows, err := q.query(ctx, q.tasksWithStatusStmt, tasksWithStatus, pq.Array(statuses))
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

const transitionTaskStatus = `-- name: TransitionTaskStatus :one
UPDATE
    tasks
SET
    status = CASE WHEN status = ANY ($1::task_status[]) THEN $2 ELSE status END,
    last_status_update = CASE WHEN status = ANY ($1::task_status[]) THEN NOW() ELSE last_status_update END
WHERE 1=1
    AND uuid=$3
RETURNING
    status
`

type TransitionTaskStatusParams struct {
	FromStatuses []TaskStatus
	ToStatus     TaskStatus
	UUID         uuid.UUID
}

func (q *Queries) TransitionTaskStatus(ctx context.Context, arg TransitionTaskStatusParams) (TaskStatus, error) {
	row := q.queryRow(ctx, q.transitionTaskStatusStmt, transitionTaskStatus, pq.Array(arg.FromStatuses), arg.ToStatus, arg.UUID)
	var status TaskStatus
	err := row.Scan(&status)
	return status, err
}

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