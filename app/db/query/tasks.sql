-- name: WorkerTasksWithStatus :many
SELECT
    *
FROM
    tasks
WHERE 1=1
    AND worker=sqlc.arg(worker)
    AND status = ANY (sqlc.arg(statuses)::task_status[])
;
