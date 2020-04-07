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
    task_status,
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
