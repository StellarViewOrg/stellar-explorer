CREATE TABLE IF NOT EXISTS live_transactions (
    profile_id TEXT NOT NULL,
    hash TEXT NOT NULL,
    ledger_sequence INTEGER NOT NULL,
    application_order INTEGER NOT NULL DEFAULT 0,
    account TEXT NOT NULL DEFAULT '',
    operation_count INTEGER NOT NULL DEFAULT 0,
    status INTEGER NOT NULL DEFAULT 0,
    is_soroban INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    cached_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (profile_id, hash)
);

CREATE INDEX IF NOT EXISTS idx_live_transactions_profile_cached
    ON live_transactions (profile_id, cached_at DESC);

CREATE INDEX IF NOT EXISTS idx_live_transactions_profile_ledger
    ON live_transactions (profile_id, ledger_sequence DESC, application_order DESC);
