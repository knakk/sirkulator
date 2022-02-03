CREATE TABLE files.image (
    id      TEXT PRIMARY KEY, -- corresponds to resource.id
    type    TEXT NOT NULL,    -- CHECK(type IN ("jpeg", "webp", "png", "svg", "avif"))
    width   INTEGER NOT NULL,
    height  INTEGER NOT NULL,
    data    BLOB NOT NULL,
    source  TEXT              -- originating URL if known
);

-- increment with 1 for each migration
PRAGMA files.user_version=1;
