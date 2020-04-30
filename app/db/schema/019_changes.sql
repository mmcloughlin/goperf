-- +goose Up
CREATE TABLE changes (
    benchmark_uuid UUID NOT NULL REFERENCES benchmarks,
    index INT NOT NULL REFERENCES commit_positions (index),

    effect_size DOUBLE PRECISION NOT NULL,

    pre_n INT NOT NULL,
    pre_mean DOUBLE PRECISION NOT NULL,
    pre_stddev DOUBLE PRECISION NOT NULL,

    post_n INT NOT NULL,
    post_mean DOUBLE PRECISION NOT NULL,
    post_stddev DOUBLE PRECISION NOT NULL,

    UNIQUE(benchmark_uuid, index)
);

-- +goose Down
DROP TABLE changes;
