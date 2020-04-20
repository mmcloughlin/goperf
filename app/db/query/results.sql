-- name: Result :one
SELECT * FROM results
WHERE uuid = $1 LIMIT 1;

-- name: BenchmarkResults :many
SELECT * FROM results
WHERE benchmark_uuid = $1;

-- name: BenchmarkPoints :many
SELECT
    r.uuid AS result_uuid,
    r.environment_uuid,
    c.sha AS commit_sha,
    c.commit_time,
    r.value
FROM results AS r
LEFT JOIN commits AS c
    ON r.commit_sha = c.sha
WHERE 1=1
    AND r.benchmark_uuid = sqlc.arg(benchmark_uuid)
    AND c.commit_time BETWEEN sqlc.arg(commit_time_start) AND sqlc.arg(commit_time_end)
ORDER BY
    c.commit_time DESC
;

-- name: InsertResult :exec
INSERT INTO results (
    uuid,
    datafile_uuid,
    line,
    benchmark_uuid,
    commit_sha,
    environment_uuid,
    metadata_uuid,
    iterations,
    value
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9
);
