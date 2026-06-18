CREATE TABLE IF NOT EXISTS list_cache (
    profile_id TEXT NOT NULL,
    list_kind TEXT NOT NULL,
    query_key TEXT NOT NULL,
    payload TEXT NOT NULL,
    source_label TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (profile_id, list_kind, query_key)
);

CREATE INDEX IF NOT EXISTS idx_list_cache_profile_updated
    ON list_cache (profile_id, updated_at DESC);