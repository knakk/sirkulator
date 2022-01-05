CREATE TABLE files.image (
    id      TEXT PRIMARY KEY,
    type    TEXT NOT NULL, -- CHECK(type IN ("jpeg", "webp", "png", "svg", "avif"))
    data    BLOB,
    source  TEXT -- originating URL if known
);

-- increment with 1 for each migration
PRAGMA files.user_version=1;
