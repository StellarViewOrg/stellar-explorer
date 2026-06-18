CREATE TABLE IF NOT EXISTS label_targets (
    id TEXT PRIMARY KEY,
    label_id TEXT NOT NULL,
    profile_id TEXT NOT NULL,
    kind TEXT NOT NULL,
    target TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(label_id, profile_id, kind, target)
);

CREATE INDEX IF NOT EXISTS idx_label_targets_label_id
    ON label_targets(label_id);

CREATE INDEX IF NOT EXISTS idx_label_targets_profile_id
    ON label_targets(profile_id);
