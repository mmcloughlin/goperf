-- +goose Up
CREATE TYPE task_status AS ENUM (
  'created',
  'in_progress',
  'complete_success',
  'complete_error'
);

CREATE TYPE task_type AS ENUM (
  'module'
);

CREATE TABLE tasks (
    uuid UUID PRIMARY KEY,
    worker TEXT NOT NULL,
    commit_sha BYTEA NOT NULL,
    type task_type NOT NULL,
    target_uuid UUID NOT NULL,
    status task_status NOT NULL,
    last_status_update TIMESTAMP WITH TIME ZONE NOT NULL,
    datafile_uuid UUID REFERENCES datafiles
);

-- +goose Down
DROP TABLE tasks;
DROP TYPE task_type;
DROP TYPE task_status;
