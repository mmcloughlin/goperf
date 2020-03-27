// Code generated by sqlc. DO NOT EDIT.
// source: modules.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const insertModule = `-- name: InsertModule :exec
INSERT INTO modules (
    uuid,
    path,
    version
) VALUES (
    $1,
    $2,
    $3
)
`

type InsertModuleParams struct {
	UUID    uuid.UUID
	Path    string
	Version string
}

func (q *Queries) InsertModule(ctx context.Context, arg InsertModuleParams) error {
	_, err := q.db.ExecContext(ctx, insertModule, arg.UUID, arg.Path, arg.Version)
	return err
}

const module = `-- name: Module :one
SELECT uuid, path, version FROM modules
WHERE uuid = $1 LIMIT 1
`

func (q *Queries) Module(ctx context.Context, uuid uuid.UUID) (Module, error) {
	row := q.db.QueryRowContext(ctx, module, uuid)
	var i Module
	err := row.Scan(&i.UUID, &i.Path, &i.Version)
	return i, err
}
