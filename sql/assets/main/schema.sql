CREATE TABLE resource (
    id          TEXT PRIMARY KEY NOT NULL, --  DEFAULT ('R' || hex(randomblob(4)) || strftime('%s','now')),
    type        TEXT NOT NULL,
    label       TEXT NOT NULL,
    gain        REAL NOT NULL DEFAULT 1.0,
    data        JSON NOT NULL DEFAULT '{}',
    created_at  INTEGER NOT NULL, -- time.Now().Unix()
    updated_at  INTEGER NOT NULL, -- time.Now().Unix()
    indexed_at  INTEGER, -- imte.Now().Unix()
    archived_at INTEGER  -- time.Now().Unix()
);

CREATE TABLE resource_edit_log (
    at          INTEGER NOT NULL, -- time.Now().Unix()
    resource_id TEXT NOT NULL REFERENCES resource (id),
    diff        JSON NOT NULL,

    PRIMARY KEY(resource_id, at)
);

CREATE TABLE relation (
    from_id  TEXT NOT NULL REFERENCES resource (id),
    to_id    TEXT NOT NULL REFERENCES resource (id),
    type     TEXT NOT NULL, -- contributes_to|subject_of|in_series|followed_by|derived_from|translation_of|
    data     JSON,

    PRIMARY KEY(from_id, to_id, type, data)
);

CREATE TABLE link (
    resource_id TEXT REFERENCES resource (id),
    type        TEXT NOT NULL, -- viaf|isbn|bibsys|wikidata
    id          TEXT NOT NULL,
    --verified_at INTEGER -- seconds since epoch

    PRIMARY KEY(resource_id, type, id)
);

CREATE INDEX idx_link_id ON link (id);

CREATE TABLE review (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    resource_id TEXT NOT NULL REFERENCES resource (id),
    prop        TEXT NOT NULL, -- or path?
    value       TEXT NOT NULL,
    data        JSON,
    queued_at   INTEGER NOT NULL -- time.Now().Unix()
);

-- increment with 1 for each migration
PRAGMA user_version=1;
