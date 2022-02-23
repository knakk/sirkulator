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
-- TODO index on expression json_extract(data, '$.year')?

CREATE TABLE resource_edit_log (
    at          INTEGER NOT NULL, -- time.Now().Unix()
    resource_id TEXT NOT NULL REFERENCES resource (id),
    diff        JSON NOT NULL,

    PRIMARY KEY(resource_id, at)
);

-- A relation describes a relation between two known resources.
--
-- If one part of the relation is a publication, it should be in the
-- originating position (from_id).
--
-- The value in the type column should be named so that the direction of the
-- relation is obvious; for example 'subject_of', rather than just 'subject',
-- where you wouldn't know if A is subject of B or B is subject of A.
CREATE TABLE relation (
    from_id  TEXT NOT NULL REFERENCES resource (id),
    to_id    TEXT NOT NULL REFERENCES resource (id),
    type     TEXT NOT NULL, -- has_contributor|has_subject|in_series|followed_by|derived_from|translation_of|
    data     JSON,

    PRIMARY KEY(from_id, to_id, type, data)
);

CREATE INDEX idx_relation_to_id ON relation (to_id);

CREATE TABLE link (
    resource_id TEXT REFERENCES resource (id),
    type        TEXT NOT NULL, -- viaf|isbn|bibsys|wikidata
    id          TEXT NOT NULL,
    --verified_at INTEGER -- seconds since epoch

    PRIMARY KEY(resource_id, type, id)
);

CREATE INDEX idx_link_id ON link (id);

-- A review is a relation where we only know the identity of
-- the resource from which the relation is pointing, but the
-- resource it is pointing to, is unknown, and must be manually
-- matched by a human by looking at the information in data.
--
-- TODO consider other nouns as name for this table; a review in the
--      context of books usually means something else..
CREATE TABLE review (
    from_id     TEXT REFERENCES resource (id),
    type        TEXT NOT NULL,
    data        JSON,
    queued_at   INTEGER NOT NULL, -- time.Now().Unix()
    -- status   TEXT, -- maybe? new|parked|etc, Status could also be stored in data column
    PRIMARY KEY(from_id, type, data)
);

-- increment with 1 for each migration
PRAGMA user_version=1;
