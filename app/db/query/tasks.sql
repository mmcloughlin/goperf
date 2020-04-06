-- name: WorkerTasksWithSpecAndStatus :many
SELECT
    *
FROM
    tasks
WHERE 1=1
    AND worker=sqlc.arg(worker)
    AND type=sqlc.arg(type)
    AND target_uuid=sqlc.arg(target_uuid)
    AND commit_sha=sqlc.arg(commit_sha)
    AND status = ANY (sqlc.arg(statuses)::task_status[])
;
