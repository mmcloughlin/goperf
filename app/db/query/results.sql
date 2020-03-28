-- name: Result :one
SELECT * FROM results
WHERE uuid = $1 LIMIT 1;

-- name: BenchmarkResults :many
SELECT * FROM results
WHERE benchmark_uuid = $1;

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
