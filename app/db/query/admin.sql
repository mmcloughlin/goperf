-- name: TruncateAll :exec
TRUNCATE
    benchmarks,
    changes,
    changes_ranked,
    commit_positions,
    commit_refs,
    commits,
    datafiles,
    modules,
    packages,
    points,
    properties,
    results,
    tasks
;
