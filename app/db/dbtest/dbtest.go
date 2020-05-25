// Package dbtest provides helpers for testing the database storage layer.
package dbtest

import (
	"context"
	"flag"
	"testing"

	"github.com/mmcloughlin/goperf/app/db"
)

var conn = flag.String("conn", "", "database connection string")

// Open a database connection.
func Open(t *testing.T) *db.DB {
	if *conn == "" {
		t.Skip("no database connection string provided")
	}

	db, err := db.Open(context.Background(), *conn)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	})

	return db
}
