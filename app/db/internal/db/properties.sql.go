// Code generated by sqlc. DO NOT EDIT.
// source: properties.sql

package db

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

const insertProperties = `-- name: InsertProperties :exec
INSERT INTO properties (
    uuid,
    fields
) VALUES (
    $1,
    $2
) ON CONFLICT DO NOTHING
`

type InsertPropertiesParams struct {
	UUID   uuid.UUID
	Fields json.RawMessage
}

func (q *Queries) InsertProperties(ctx context.Context, arg InsertPropertiesParams) error {
	_, err := q.db.ExecContext(ctx, insertProperties, arg.UUID, arg.Fields)
	return err
}

const properties = `-- name: Properties :one
SELECT uuid, fields FROM properties
WHERE uuid = $1 LIMIT 1
`

func (q *Queries) Properties(ctx context.Context, uuid uuid.UUID) (Property, error) {
	row := q.db.QueryRowContext(ctx, properties, uuid)
	var i Property
	err := row.Scan(&i.UUID, &i.Fields)
	return i, err
}
