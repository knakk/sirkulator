CREATE TABLE resource (
    id          TEXT PRIMARY KEY NOT NULL, --  DEFAULT ('R' || hex(randomblob(4)) || strftime('%s','now')),
    type        TEXT NOT NULL,
    label       TEXT NOT NULL,
    gain        REAL NOT NULL DEFAULT 1.0,
    data        JSON NOT NULL DEFAULT '{}',
    created_at  INTEGER NOT NULL, -- time.Now().Unix()
    updated_at  INTEGER NOT NULL, -- time.Now().Unix()
    archived_at INTEGER  -- time.Now().Unix()
);
-- TODO index on expression json_extract(data, '$.year')?

CREATE TABLE resource_edit_log (
    at          INTEGER NOT NULL, -- time.Now().Unix()
    resource_id TEXT NOT NULL REFERENCES resource (id),
    diff        JSON NOT NULL,

    PRIMARY KEY(resource_id, at)
);

-- A relation describes a relation between two resources.
--
-- If to_id is NULL, the relation is considered a review which
-- must be manually resolved. A review should have some info
-- in `data` field to help with that - at least a `label` key.
--
-- If one part of the relation is a publication, it should be in the
-- originating position (from_id).
--
-- The value in the type column should be named so that the direction of the
-- relation is obvious; for example 'subject_of', rather than just 'subject',
-- where you wouldn't know if A is subject of B or B is subject of A.
CREATE TABLE relation (
    id        INTEGER PRIMARY KEY,
    from_id   TEXT NOT NULL REFERENCES resource (id),
    to_id     TEXT REFERENCES resource (id),
    type      TEXT NOT NULL, -- has_contributor|has_subject|in_series|followed_by|derived_from|translation_of|
    data      JSON,
    queued_at INTEGER -- Only set if to_id is NULL
);

CREATE INDEX idx_relation_from_id ON relation (from_id);
CREATE INDEX idx_relation_to_id ON relation (to_id);

CREATE TABLE link (
    resource_id TEXT REFERENCES resource (id),
    type        TEXT NOT NULL, -- viaf|isbn|bibsys|wikidata
    id          TEXT NOT NULL,
    --verified_at INTEGER -- seconds since epoch

    PRIMARY KEY(resource_id, type, id)
);

CREATE INDEX idx_link_id ON link (id);

CREATE TABLE job_run (
    id       INTEGER PRIMARY KEY,
    name     TEXT NOT NULL,
    start_at INTEGER NOT NULL, -- time.Now().Unix()
    stop_at  INTEGER,          -- time.Now().Unix()
    status   TEXT NOT NULL,    -- running|done|failed|cancelled
    output   BLOB -- gzipped text
);

CREATE TABLE job_schedule (
    id       INTEGER PRIMARY KEY,
    name     TEXT NOT NULL,
    cron     TEXT NOT NULL -- cron expression with seconds precision
);

-- increment with 1 for each migration
PRAGMA user_version=1;
