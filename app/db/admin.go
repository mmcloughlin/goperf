package db

import (
	"context"

	"github.com/mmcloughlin/goperf/app/db/internal/db"
)

// TruncateAll deletes all data from the database.
func (d *DB) TruncateAll(ctx context.Context) error {
	return d.txq(ctx, func(q *db.Queries) error {
		return q.TruncateAll(ctx)
	})
}
