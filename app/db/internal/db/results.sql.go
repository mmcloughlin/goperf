// Code generated by sqlc. DO NOT EDIT.
// source: results.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const benchmarkPoints = `-- name: BenchmarkPoints :many
SELECT
    result_uuid,
    environment_uuid,
    commit_sha,
    commit_index,
    value
FROM
    points
WHERE 1=1
    AND benchmark_uuid = $1
    AND commit_index BETWEEN $2 AND $3
ORDER BY
    commit_index
`

type BenchmarkPointsParams struct {
	BenchmarkUUID  uuid.UUID
	CommitIndexMin int32
	CommitIndexMax int32
}

type BenchmarkPointsRow struct {
	ResultUUID      uuid.UUID
	EnvironmentUUID uuid.UUID
	CommitSHA       []byte
	CommitIndex     int32
	Value           float64
}

func (q *Queries) BenchmarkPoints(ctx context.Context, arg BenchmarkPointsParams) ([]BenchmarkPointsRow, error) {
	rows, err := q.query(ctx, q.benchmarkPointsStmt, benchmarkPoints, arg.BenchmarkUUID, arg.CommitIndexMin, arg.CommitIndexMax)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []BenchmarkPointsRow
	for rows.Next() {
		var i BenchmarkPointsRow
		if err := rows.Scan(
			&i.ResultUUID,
			&i.EnvironmentUUID,
			&i.CommitSHA,
			&i.CommitIndex,
			&i.Value,
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

const benchmarkResults = `-- name: BenchmarkResults :many
SELECT uuid, datafile_uuid, line, benchmark_uuid, commit_sha, environment_uuid, metadata_uuid, iterations, value FROM results
WHERE benchmark_uuid = $1
`

func (q *Queries) BenchmarkResults(ctx context.Context, benchmarkUuid uuid.UUID) ([]Result, error) {
	rows, err := q.query(ctx, q.benchmarkResultsStmt, benchmarkResults, benchmarkUuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Result
	for rows.Next() {
		var i Result
		if err := rows.Scan(
			&i.UUID,
			&i.DatafileUUID,
			&i.Line,
			&i.BenchmarkUUID,
			&i.CommitSHA,
			&i.EnvironmentUUID,
			&i.MetadataUUID,
			&i.Iterations,
			&i.Value,
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

const insertResult = `-- name: InsertResult :exec
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
)
`

type InsertResultParams struct {
	UUID            uuid.UUID
	DatafileUUID    uuid.UUID
	Line            int32
	BenchmarkUUID   uuid.UUID
	CommitSHA       []byte
	EnvironmentUUID uuid.UUID
	MetadataUUID    uuid.UUID
	Iterations      int64
	Value           float64
}

func (q *Queries) InsertResult(ctx context.Context, arg InsertResultParams) error {
	_, err := q.exec(ctx, q.insertResultStmt, insertResult,
		arg.UUID,
		arg.DatafileUUID,
		arg.Line,
		arg.BenchmarkUUID,
		arg.CommitSHA,
		arg.EnvironmentUUID,
		arg.MetadataUUID,
		arg.Iterations,
		arg.Value,
	)
	return err
}

const result = `-- name: Result :one
SELECT uuid, datafile_uuid, line, benchmark_uuid, commit_sha, environment_uuid, metadata_uuid, iterations, value FROM results
WHERE uuid = $1 LIMIT 1
`

func (q *Queries) Result(ctx context.Context, uuid uuid.UUID) (Result, error) {
	row := q.queryRow(ctx, q.resultStmt, result, uuid)
	var i Result
	err := row.Scan(
		&i.UUID,
		&i.DatafileUUID,
		&i.Line,
		&i.BenchmarkUUID,
		&i.CommitSHA,
		&i.EnvironmentUUID,
		&i.MetadataUUID,
		&i.Iterations,
		&i.Value,
	)
	return i, err
}

const tracePoints = `-- name: TracePoints :many
SELECT
    benchmark_uuid,
    environment_uuid,
    commit_index,
    value
FROM
    points
WHERE 1=1
    AND commit_index BETWEEN $1 AND $2
`

type TracePointsParams struct {
	CommitIndexMin int32
	CommitIndexMax int32
}

type TracePointsRow struct {
	BenchmarkUUID   uuid.UUID
	EnvironmentUUID uuid.UUID
	CommitIndex     int32
	Value           float64
}

func (q *Queries) TracePoints(ctx context.Context, arg TracePointsParams) ([]TracePointsRow, error) {
	rows, err := q.query(ctx, q.tracePointsStmt, tracePoints, arg.CommitIndexMin, arg.CommitIndexMax)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TracePointsRow
	for rows.Next() {
		var i TracePointsRow
		if err := rows.Scan(
			&i.BenchmarkUUID,
			&i.EnvironmentUUID,
			&i.CommitIndex,
			&i.Value,
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
