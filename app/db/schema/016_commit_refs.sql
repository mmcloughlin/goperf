-- +goose Up
CREATE TABLE commit_refs (
    sha BYTEA REFERENCES commits,
    ref TEXT NOT NULL,
    UNIQUE(sha, ref)
);

-- +goose Down
DROP TABLE commit_refs;
