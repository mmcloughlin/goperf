-- +goose Up
CREATE TABLE modules (
    uuid UUID PRIMARY KEY,
    path TEXT NOT NULL,
    version TEXT
);

-- +goose Down
DROP TABLE modules;
