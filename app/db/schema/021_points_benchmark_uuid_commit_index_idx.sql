-- +goose Up
CREATE INDEX points_benchmark_uuid_commit_index_idx ON points (benchmark_uuid, commit_index);

-- +goose Down
DROP INDEX points_benchmark_uuid_commit_index_idx;
