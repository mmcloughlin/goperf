-- +goose Up
CREATE TABLE modules (
    uuid UUID PRIMARY KEY,
    path TEXT NOT NULL,
    version TEXT NOT NULL
);

-- +goose Down
DROP TABLE modules;
