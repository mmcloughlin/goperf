// Code generated by sqlc. DO NOT EDIT.
// source: changes.sql

package db

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

const buildChangesRanked = `-- name: BuildChangesRanked :exec
INSERT INTO changes_ranked (
    SELECT
        benchmark_uuid, environment_uuid, commit_index, effect_size, pre_n, pre_mean, pre_stddev, post_n, post_mean, post_stddev,
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
`

func (q *Queries) BuildChangesRanked(ctx context.Context) error {
	_, err := q.exec(ctx, q.buildChangesRankedStmt, buildChangesRanked)
	return err
}

const changeSummaries = `-- name: ChangeSummaries :many
SELECT
    chg.benchmark_uuid, chg.environment_uuid, chg.commit_index, chg.effect_size, chg.pre_n, chg.pre_mean, chg.pre_stddev, chg.post_n, chg.post_mean, chg.post_stddev, chg.rank_by_effect_size, chg.rank_by_abs_percent_change,
    c.sha AS commit_sha,
    SPLIT_PART(c.message, E'\n', 1)::TEXT AS commit_subject,

    b.uuid, b.package_uuid, b.full_name, b.name, b.unit, b.parameters,
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
    AND ABS(chg.effect_size) > $1
    AND chg.commit_index BETWEEN $2 AND $3
    AND chg.rank_by_effect_size <= $4
    AND chg.rank_by_abs_percent_change <= $5
ORDER BY
    commit_index DESC
`

type ChangeSummariesParams struct {
	EffectSizeMin             float64
	CommitIndexMin            int32
	CommitIndexMax            int32
	RankByEffectSizeMax       int32
	RankByAbsPercentChangeMax int32
}

type ChangeSummariesRow struct {
	BenchmarkUUID          uuid.UUID
	EnvironmentUUID        uuid.UUID
	CommitIndex            int32
	EffectSize             float64
	PreN                   int32
	PreMean                float64
	PreStddev              float64
	PostN                  int32
	PostMean               float64
	PostStddev             float64
	RankByEffectSize       int32
	RankByAbsPercentChange int32
	CommitSHA              []byte
	CommitSubject          string
	UUID                   uuid.UUID
	PackageUUID            uuid.UUID
	FullName               string
	Name                   string
	Unit                   string
	Parameters             json.RawMessage
	RelativePath           string
	Path                   string
	Version                string
}

func (q *Queries) ChangeSummaries(ctx context.Context, arg ChangeSummariesParams) ([]ChangeSummariesRow, error) {
	rows, err := q.query(ctx, q.changeSummariesStmt, changeSummaries,
		arg.EffectSizeMin,
		arg.CommitIndexMin,
		arg.CommitIndexMax,
		arg.RankByEffectSizeMax,
		arg.RankByAbsPercentChangeMax,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ChangeSummariesRow
	for rows.Next() {
		var i ChangeSummariesRow
		if err := rows.Scan(
			&i.BenchmarkUUID,
			&i.EnvironmentUUID,
			&i.CommitIndex,
			&i.EffectSize,
			&i.PreN,
			&i.PreMean,
			&i.PreStddev,
			&i.PostN,
			&i.PostMean,
			&i.PostStddev,
			&i.RankByEffectSize,
			&i.RankByAbsPercentChange,
			&i.CommitSHA,
			&i.CommitSubject,
			&i.UUID,
			&i.PackageUUID,
			&i.FullName,
			&i.Name,
			&i.Unit,
			&i.Parameters,
			&i.RelativePath,
			&i.Path,
			&i.Version,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const deleteChangesCommitRange = `-- name: DeleteChangesCommitRange :exec
DELETE FROM changes
WHERE 1=1
    AND commit_index BETWEEN $1 AND $2
`

type DeleteChangesCommitRangeParams struct {
	CommitIndexMin int32
	CommitIndexMax int32
}

func (q *Queries) DeleteChangesCommitRange(ctx context.Context, arg DeleteChangesCommitRangeParams) error {
	_, err := q.exec(ctx, q.deleteChangesCommitRangeStmt, deleteChangesCommitRange, arg.CommitIndexMin, arg.CommitIndexMax)
	return err
}
