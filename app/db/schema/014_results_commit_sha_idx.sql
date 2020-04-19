-- +goose Up
CREATE INDEX results_commit_sha_idx ON results (commit_sha);

-- +goose Down
DROP INDEX results_commit_sha_idx;
