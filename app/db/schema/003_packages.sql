-- +goose Up
CREATE TABLE packages (
    uuid UUID PRIMARY KEY,
    module_uuid UUID NOT NULL REFERENCES modules,
    relative_path TEXT NOT NULL
);

-- +goose Down
DROP TABLE packages;
