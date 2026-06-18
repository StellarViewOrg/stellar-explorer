CREATE TABLE IF NOT EXISTS saved_views (
    id TEXT PRIMARY KEY,
    profile_id TEXT NOT NULL,
    name TEXT NOT NULL,
    command TEXT NOT NULL,
    screen TEXT NOT NULL DEFAULT '',
    entity_kind TEXT NOT NULL DEFAULT '',
    entity_target TEXT NOT NULL DEFAULT '',
    filters_json TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(profile_id, name)
);

CREATE INDEX IF NOT EXISTS idx_saved_views_profile_updated
    ON saved_views(profile_id, updated_at DESC);