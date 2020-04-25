-- +goose Up
CREATE TABLE commit_positions (
    sha BYTEA PRIMARY KEY REFERENCES commits,
    commit_time TIMESTAMP WITH TIME ZONE NOT NULL,
    index INT NOT NULL
);

-- +goose Down
DROP TABLE commit_positions;
