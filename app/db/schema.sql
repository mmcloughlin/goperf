CREATE TABLE IF NOT EXISTS commits (
    sha BYTEA PRIMARY KEY,
    tree BYTEA,
    parents BYTEA[],
    author_name TEXT,
    author_email TEXT,
    author_time TIMESTAMP WITH TIME ZONE,
    committer_name TEXT,
    committer_email TEXT,
    commit_time TIMESTAMP WITH TIME ZONE,
    message TEXT
);

CREATE TABLE IF NOT EXISTS modules (
    uuid UUID PRIMARY KEY,
    path TEXT NOT NULL,
    version TEXT
);

CREATE TABLE IF NOT EXISTS packages (
    uuid UUID PRIMARY KEY,
    module_uuid UUID NOT NULL REFERENCES modules,
    relative_path TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS benchmarks (
    uuid UUID PRIMARY KEY,
    package_uuid UUID NOT NULL REFERENCES packages,
    full_name TEXT NOT NULL,
    name TEXT NOT NULL,
    unit TEXT NOT NULL,
    parameters JSONB
);

CREATE TABLE IF NOT EXISTS datafiles (
    uuid UUID PRIMARY KEY,
    name TEXT NOT NULL,
    sha256 BYTEA NOT NULL
);

CREATE TABLE IF NOT EXISTS properties (
    uuid UUID PRIMARY KEY,
    fields JSONB
);

CREATE TABLE IF NOT EXISTS results (
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
