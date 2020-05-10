-- name: Result :one
SELECT * FROM results
WHERE uuid = $1 LIMIT 1;

-- name: BenchmarkResults :many
SELECT * FROM results
WHERE benchmark_uuid = $1;

-- name: BenchmarkPoints :many
SELECT
    result_uuid,
    environment_uuid,
    commit_sha,
    commit_index,
    value
FROM
    points
WHERE 1=1
    AND benchmark_uuid = sqlc.arg(benchmark_uuid)
    AND commit_index BETWEEN sqlc.arg(commit_index_min) AND sqlc.arg(commit_index_max)
ORDER BY
    commit_index
;

-- name: TracePoints :many
SELECT
    benchmark_uuid,
    environment_uuid,
    commit_index,
    value
FROM
    points
WHERE 1=1
    AND commit_index BETWEEN sqlc.arg(commit_index_min) AND sqlc.arg(commit_index_max)
;

-- name: Trace :many
SELECT
    commit_index,
    value
FROM
    points
WHERE 1=1
    AND benchmark_uuid = sqlc.arg(benchmark_uuid)
    AND environment_uuid = sqlc.arg(environment_uuid)
    AND commit_index BETWEEN sqlc.arg(commit_index_min) AND sqlc.arg(commit_index_max)
ORDER BY
    commit_index
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
