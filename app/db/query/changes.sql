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
    changes_ranked AS chg
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
    AND chg.commit_index BETWEEN sqlc.arg(commit_index_min) AND sqlc.arg(commit_index_max)
    AND chg.rank_by_effect_size <= sqlc.arg(rank_by_effect_size_max)
    AND chg.rank_by_abs_percent_change <= sqlc.arg(rank_by_abs_percent_change_max)
ORDER BY
    commit_index DESC
;

-- name: BuildChangesRanked :exec
INSERT INTO changes_ranked (
    SELECT
        *,
        ROW_NUMBER() OVER (
            PARTITION BY commit_index
            ORDER BY ABS(effect_size) DESC
        ) AS rank_by_effect_size,
        ROW_NUMBER() OVER (
            PARTITION BY commit_index
            ORDER BY ABS((post_mean/pre_mean)-1.0) DESC
        ) AS rank_by_abs_percent_change
    FROM
        changes
)
ON CONFLICT (benchmark_uuid, environment_uuid, commit_index)
DO UPDATE SET
    rank_by_effect_size = EXCLUDED.rank_by_effect_size,
    rank_by_abs_percent_change = EXCLUDED.rank_by_abs_percent_change
;
