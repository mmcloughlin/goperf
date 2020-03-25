-- +goose Up
CREATE TABLE results (
    uuid UUID PRIMARY KEY,
    datafile_uuid UUID NOT NULL REFERENCES datafiles,
    line INTEGER NOT NULL,
    benchmark_uuid UUID NOT NULL REFERENCES benchmarks,
    commit_sha BYTEA NOT NULL REFERENCES commits,
    environment_uuid UUID NOT NULL REFERENCES properties,
    metadata_uuid UUID NOT NULL REFERENCES properties,
    iterations BIGINT,
    value DOUBLE PRECISION NOT NULL
);

-- +goose Down
DROP TABLE results;
