// Code generated by sqlc. DO NOT EDIT.
// source: benchmarks.sql

package db

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

const benchmark = `-- name: Benchmark :one
SELECT uuid, package_uuid, full_name, name, unit, parameters FROM benchmarks
WHERE uuid = $1 LIMIT 1
`

func (q *Queries) Benchmark(ctx context.Context, uuid uuid.UUID) (Benchmark, error) {
	row := q.db.QueryRowContext(ctx, benchmark, uuid)
	var i Benchmark
	err := row.Scan(
		&i.UUID,
		&i.PackageUUID,
		&i.FullName,
		&i.Name,
		&i.Unit,
		&i.Parameters,
	)
	return i, err
}

const insertBenchmark = `-- name: InsertBenchmark :exec
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
) ON CONFLICT DO NOTHING
`

type InsertBenchmarkParams struct {
	UUID        uuid.UUID
	PackageUUID uuid.UUID
	FullName    string
	Name        string
	Unit        string
	Parameters  json.RawMessage
}

func (q *Queries) InsertBenchmark(ctx context.Context, arg InsertBenchmarkParams) error {
	_, err := q.db.ExecContext(ctx, insertBenchmark,
		arg.UUID,
		arg.PackageUUID,
		arg.FullName,
		arg.Name,
		arg.Unit,
		arg.Parameters,
	)
	return err
}

const packageBenchmarks = `-- name: PackageBenchmarks :many
SELECT uuid, package_uuid, full_name, name, unit, parameters FROM benchmarks
WHERE package_uuid = $1
`

func (q *Queries) PackageBenchmarks(ctx context.Context, packageUuid uuid.UUID) ([]Benchmark, error) {
	rows, err := q.db.QueryContext(ctx, packageBenchmarks, packageUuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Benchmark
	for rows.Next() {
		var i Benchmark
		if err := rows.Scan(
			&i.UUID,
			&i.PackageUUID,
			&i.FullName,
			&i.Name,
			&i.Unit,
			&i.Parameters,
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
