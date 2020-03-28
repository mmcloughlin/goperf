-- name: Benchmark :one
SELECT * FROM benchmarks
WHERE uuid = $1 LIMIT 1;

-- name: InsertBenchmark :exec
INSERT INTO benchmarks (
    uuid,
    package_uuid,
    full_name,
    name,
    unit,
    parameters
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
) ON CONFLICT DO NOTHING;
