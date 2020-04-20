-- +goose Up
CREATE INDEX commits_commit_time_idx ON commits (commit_time);

-- +goose Down
DROP INDEX commits_commit_time_idx;
