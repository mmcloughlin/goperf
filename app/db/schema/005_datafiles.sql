-- +goose Up
CREATE TABLE datafiles (
    uuid UUID PRIMARY KEY,
    name TEXT NOT NULL,
    sha256 BYTEA NOT NULL
);

-- +goose Down
DROP TABLE datafiles;
