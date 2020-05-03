-- name: DeleteChangesCommitRange :exec
DELETE FROM changes
WHERE 1=1
    AND commit_index BETWEEN sqlc.arg(commit_index_min) AND sqlc.arg(commit_index_max)
;

-- name: ChangeSummaries :many
SELECT
    chg.*,
    c.sha AS commit_sha,
    SPLIT_PART(c.message, E'\n', 1)::TEXT AS commit_subject,

    b.*,
    pkg.relative_path,
    mod.path,
    mod.version
FROM
    changes AS chg
    INNER JOIN commit_positions AS p
        ON chg.commit_index=p.index
    INNER JOIN commits AS c
        ON p.sha=c.sha
    INNER JOIN benchmarks AS b
        ON chg.benchmark_uuid=b.uuid
    INNER JOIN packages AS pkg
        ON b.package_uuid=pkg.uuid
    INNER JOIN modules AS mod
        ON pkg.module_uuid=mod.uuid
WHERE 1=1
    AND ABS(chg.effect_size) > sqlc.arg(effect_size_min)
    AND commit_index BETWEEN sqlc.arg(commit_index_min) AND sqlc.arg(commit_index_max)
ORDER BY
    commit_index DESC
;
