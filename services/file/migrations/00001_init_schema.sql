-- +goose Up
CREATE TABLE IF NOT EXISTS files (
    id       TEXT PRIMARY KEY,
    user_id  TEXT NOT NULL,
    filename TEXT NOT NULL,
    status   TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS mem_metadata (
    file_id      TEXT PRIMARY KEY REFERENCES files(id),
    pid          INTEGER,
    tid          INTEGER,
    heap_addr    BIGINT,
    heap_size    BIGINT,
    stack_offset BIGINT,
    fd           INTEGER,
    nr_mmap      INTEGER,
    nr_write     INTEGER,
    nr_fsync     INTEGER,
    nr_openat    INTEGER,
    captured_at  TEXT
);

-- +goose Down
DROP TABLE IF EXISTS mem_metadata;
DROP TABLE IF EXISTS files;
