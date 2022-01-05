CREATE TABLE job_run (
    id          INTEGER     PRIMARY KEY AUTOINCREMENT,
    name        TEXT        NOT NULL,
    started_at  INTEGER     NOT NULL,
    finished_at INTEGER     CHECK(finished_at IS NOT NULL OR status="running"),
    status      TEXT        NOT NULL CHECK(status IN ("running", "cancelled", "completed", "failed")),
    output      BLOB        -- gzipped
);

-- increment with 1 for each migration
PRAGMA user_version=1;
