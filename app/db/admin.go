package db

import (
	"context"

	"github.com/mmcloughlin/cb/app/db/internal/db"
)

// TruncateNonStatic deletes almost all data from the database. Only static
// tables such as commits and modules are preserved.
func (d *DB) TruncateNonStatic(ctx context.Context) error {
	return d.tx(ctx, func(q *db.Queries) error {
		return q.TruncateNonStatic(ctx)
	})
}
