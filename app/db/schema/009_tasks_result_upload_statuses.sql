-- +goose NO TRANSACTION

-- +goose Up
ALTER TYPE task_status ADD VALUE 'result_upload_started';
ALTER TYPE task_status ADD VALUE 'result_uploaded';
