package db

import (
	"database/sql"

	"github.com/mmcloughlin/cb/app/db/internal/db"
)

//go:generate sqlc generate

type DB struct {
	db *sql.DB
	q  *db.Queries
}

func Open(conn string) (*DB, error) {
	d, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	return &DB{
		db: d,
		q:  db.New(d),
	}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}
