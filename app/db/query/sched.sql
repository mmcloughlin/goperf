-- name: RecentCommitModulePairsWithoutWorkerResults :many
SELECT
    c.sha AS commit_sha,
    m.uuid AS module_uuid
FROM
    commits AS c,
    modules AS m
WHERE 1=1
    AND c.commit_time > sqlc.arg(since)
    AND NOT EXISTS (
        SELECT *
        FROM
            results AS r
            LEFT JOIN benchmarks AS b ON r.benchmark_uuid = b.uuid
            LEFT JOIN packages AS p ON b.package_uuid = p.uuid
            LEFT JOIN datafiles AS f ON r.datafile_uuid = f.uuid
            LEFT JOIN tasks AS t ON f.uuid = t.datafile_uuid
        WHERE 1=1
            AND r.commit_sha = c.sha
            AND p.module_uuid = m.uuid
            AND t.worker = sqlc.arg(worker)
    )
ORDER BY
    c.commit_time DESC,
    m.uuid
LIMIT
    sqlc.arg(num)
;
