-- +goose Up
CREATE TABLE commits (
    sha BYTEA PRIMARY KEY,
    tree BYTEA,
    parents BYTEA[],
    author_name TEXT NOT NULL,
    author_email TEXT NOT NULL,
    author_time TIMESTAMP WITH TIME ZONE NOT NULL,
    committer_name TEXT NOT NULL,
    committer_email TEXT NOT NULL,
    commit_time TIMESTAMP WITH TIME ZONE NOT NULL,
    message TEXT NOT NULL
);

-- +goose Down
DROP TABLE commits;
