-- +goose NO TRANSACTION

-- +goose Up
ALTER TYPE task_status ADD VALUE 'stale_timeout';
