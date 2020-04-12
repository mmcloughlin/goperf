-- name: RecentCommitModulePairsWithoutWorkerTasks :many
SELECT
    c.sha AS commit_sha,
    c.commit_time,
    m.uuid AS module_uuid
FROM
    commits AS c,
    modules AS m
WHERE NOT EXISTS (
        SELECT *
        FROM tasks AS t
        WHERE 1=1
            AND t.commit_sha = c.sha
            AND t.type = 'module'
            AND t.target_uuid = m.uuid
            AND t.status = ANY (sqlc.arg(statuses)::task_status[])
            AND t.worker = sqlc.arg(worker)
    )
ORDER BY
    c.commit_time DESC,
    m.uuid
LIMIT
    sqlc.arg(num)
;
