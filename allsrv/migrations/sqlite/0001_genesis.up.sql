CREATE TABLE foos
(
    id         TEXT PRIMARY KEY,
    name       TEXT UNIQUE,
    note       TEXT,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
);