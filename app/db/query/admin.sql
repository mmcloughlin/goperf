-- name: TruncateNonStatic :exec
-- TruncateNonStatic is a destructive query that deletes everything apart from
-- commits and modules.
TRUNCATE
    benchmarks,
    datafiles,
    packages,
    properties,
    results,
    tasks
;
