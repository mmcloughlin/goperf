-- +goose Up
CREATE TABLE benchmarks (
    uuid UUID PRIMARY KEY,
    package_uuid UUID NOT NULL REFERENCES packages,
    full_name TEXT NOT NULL,
    name TEXT NOT NULL,
    unit TEXT NOT NULL,
    parameters JSONB
);

-- +goose Down
DROP TABLE benchmarks;
