-- +goose Up
CREATE TABLE properties (
    uuid UUID PRIMARY KEY,
    fields JSONB
);

-- +goose Down
DROP TABLE properties;
