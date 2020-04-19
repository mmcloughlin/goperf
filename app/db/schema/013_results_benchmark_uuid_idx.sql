-- +goose Up
CREATE INDEX results_benchmark_uuid_idx ON results (benchmark_uuid);

-- +goose Down
DROP INDEX results_benchmark_uuid_idx;
