-- name: RecentCommitModulePairsWithoutWorkerTasks :many
SELECT
    p.sha AS commit_sha,
    p.commit_time,
    m.uuid AS module_uuid
FROM
    commit_positions AS p,
    modules AS m
WHERE NOT EXISTS (
        SELECT *
        FROM tasks AS t
        WHERE 1=1
            AND t.commit_sha = p.sha
            AND t.type = 'module'
            AND t.target_uuid = m.uuid
            AND t.status = ANY (sqlc.arg(statuses)::task_status[])
            AND t.worker = sqlc.arg(worker)
    )
ORDER BY
    p.commit_time DESC,
    m.uuid
LIMIT
    sqlc.arg(num)
;

-- name: CommitModuleWorkerErrors :many
SELECT
    target_uuid AS module_uuid,
    commit_sha,
    COUNT(*) FILTER (WHERE status = 'complete_error') AS num_errors,
    MAX(last_status_update)::TIMESTAMP WITH TIME ZONE AS last_attempt_time
FROM
    tasks
WHERE 1=1
    AND worker = sqlc.arg(worker)
    AND type = 'module'
GROUP BY
    1, 2
HAVING 1=1
    AND COUNT(*) FILTER (WHERE status = 'complete_success') = 0
    AND COUNT(*) FILTER (WHERE status = 'complete_error') BETWEEN 1 AND sqlc.arg(max_errors)::INT
    AND MAX(last_status_update) < sqlc.arg(last_attempt_before)
ORDER BY
    num_errors ASC,
    last_attempt_time ASC
LIMIT
    sqlc.arg(num)
;
