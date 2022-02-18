
CREATE TABLE oai.source (
    id         TEXT     NOT NULL PRIMARY KEY,
    url        TEXT     NOT NULL,
    dataset    TEXT     NOT NULL,
    prefix     TEXT     NOT NULL,
    token      TEXT     NOT NULL DEFAULT "",
    in_sync_at INTEGER   -- seconds since epoch
);

CREATE TABLE oai.record (
    source_id  TEXT    NOT NULL REFERENCES source(id),
    id         TEXT    NOT NULL,
    data       BLOB    NOT NULL, -- gzipped XML oai record
    new_data   BLOB,             -- gzipped XML oai record

    created_at  INTEGER     NOT NULL, -- seconds since epoch, remote timestamp
    updated_at  INTEGER     NOT NULL, -- seconds since epoch, remote timestamp
    archived_at INTEGER,              -- seconds since epoch, remote timestamp
    queued_at   INTEGER,              -- seconds since epoch, local timestamp at moment of row insertion/update

    PRIMARY KEY (source_id, id)
);

CREATE TABLE oai.link (
    source_id   TEXT     NOT NULL REFERENCES record(source_id) ON DELETE CASCADE,
    record_id   TEXT     NOT NULL REFERENCES record(id) ON DELETE CASCADE,
    type        TEXT     NOT NULL, -- isbn|issn|isni|viaf|orcid etc
    id          TEXT     NOT NULL,

    PRIMARY KEY (source_id, record_id, id) -- TODO what about type?
);

CREATE INDEX oai.idx_id ON link (id);

-- increment with 1 for each migration
PRAGMA oai.user_version=1;
