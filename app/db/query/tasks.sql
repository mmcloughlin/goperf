-- name: Task :one
SELECT * FROM tasks
WHERE uuid = $1 LIMIT 1;

-- name: TasksWithStatus :many
SELECT
    *
FROM
    tasks
WHERE 1=1
    AND status = ANY (sqlc.arg(statuses)::task_status[])
;

-- name: WorkerTasksWithStatus :many
SELECT
    *
FROM
    tasks
WHERE 1=1
    AND worker=sqlc.arg(worker)
    AND status = ANY (sqlc.arg(statuses)::task_status[])
;

-- name: CreateTask :one
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
RETURNING *
;

-- name: TransitionTaskStatus :one
UPDATE
    tasks
SET
    status = CASE WHEN status = ANY (sqlc.arg(from_statuses)::task_status[]) THEN sqlc.arg(to_status) ELSE status END,
    last_status_update = CASE WHEN status = ANY (sqlc.arg(from_statuses)::task_status[]) THEN NOW() ELSE last_status_update END
WHERE 1=1
    AND uuid=sqlc.arg(uuid)
RETURNING
    status
;

-- name: TransitionTaskStatusesBefore :exec
UPDATE
    tasks
SET
    status = sqlc.arg(to_status),
    last_status_update = NOW()
WHERE 1=1
    AND status = ANY (sqlc.arg(from_statuses)::task_status[])
    AND last_status_update < sqlc.arg(until)
;

-- name: SetTaskDataFile :exec
UPDATE
    tasks
SET
    datafile_uuid = $1
WHERE
    uuid = $2
;
