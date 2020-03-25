-- +goose Up
CREATE TABLE commits (
    sha BYTEA PRIMARY KEY,
    tree BYTEA,
    parents BYTEA[],
    author_name TEXT,
    author_email TEXT,
    author_time TIMESTAMP WITH TIME ZONE,
    committer_name TEXT,
    committer_email TEXT,
    commit_time TIMESTAMP WITH TIME ZONE,
    message TEXT
);

-- +goose Down
DROP TABLE commits;
