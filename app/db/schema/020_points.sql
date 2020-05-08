-- +goose Up
CREATE TABLE points (
    result_uuid UUID NOT NULL PRIMARY KEY REFERENCES results,
    benchmark_uuid UUID NOT NULL REFERENCES benchmarks,
    environment_uuid UUID NOT NULL REFERENCES properties,
    commit_sha BYTEA NOT NULL REFERENCES commits,
    commit_index INT NOT NULL REFERENCES commit_positions (index),
    value DOUBLE PRECISION NOT NULL
);

-- +goose Down
DROP TABLE points;
