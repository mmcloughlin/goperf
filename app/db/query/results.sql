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
    p.index AS commit_index,
    r.value
FROM
    results AS r
    LEFT JOIN commits AS c
        ON r.commit_sha = c.sha
    INNER JOIN commit_positions AS p
        ON p.sha = c.sha
WHERE 1=1
    AND r.benchmark_uuid = sqlc.arg(benchmark_uuid)
    AND p.index BETWEEN sqlc.arg(commit_index_min) AND sqlc.arg(commit_index_max)
ORDER BY
    p.index
;

-- name: TracePoints :many
SELECT
    r.benchmark_uuid,
    r.environment_uuid,
    p.index AS commit_index,
    p.commit_time,
    r.value
FROM
    results AS r
    INNER JOIN commit_positions AS p
        ON r.commit_sha=p.sha
WHERE 1=1
    AND p.index BETWEEN sqlc.arg(commit_index_min) AND sqlc.arg(commit_index_max)
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
