-- +goose Up
ALTER TABLE commit_positions ADD CONSTRAINT commit_positions_index_key UNIQUE (index);

-- +goose Down
ALTER TABLE commit_positions DROP CONSTRAINT commit_positions_index_key;
