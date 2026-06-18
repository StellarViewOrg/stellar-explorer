CREATE TABLE IF NOT EXISTS entity_cache (
    profile_id TEXT NOT NULL,
    kind TEXT NOT NULL,
    target TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    summary TEXT NOT NULL DEFAULT '',
    payload TEXT NOT NULL,
    source_label TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    PRIMARY KEY (profile_id, kind, target)
);

CREATE INDEX IF NOT EXISTS idx_entity_cache_profile_updated
    ON entity_cache (profile_id, updated_at DESC);
