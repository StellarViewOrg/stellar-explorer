CREATE TABLE known_accounts (
    address             VARCHAR(56) PRIMARY KEY,
    label               VARCHAR(256) NOT NULL,
    category            VARCHAR(64),
    domain              VARCHAR(256),
    description         TEXT,
    tags                TEXT[],
    verified            BOOLEAN NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_known_category ON known_accounts (category);
CREATE INDEX idx_known_label ON known_accounts USING GIN (to_tsvector('english', label));

CREATE TABLE liquidity_pools (
    pool_id             CHAR(64) PRIMARY KEY,
    fee_bp              INTEGER NOT NULL,
    asset_a_type        SMALLINT NOT NULL,
    asset_a_code        VARCHAR(12),
    asset_a_issuer      VARCHAR(56),
    reserve_a           NUMERIC(20, 7) NOT NULL DEFAULT 0,
    asset_b_type        SMALLINT NOT NULL,
    asset_b_code        VARCHAR(12),
    asset_b_issuer      VARCHAR(56),
    reserve_b           NUMERIC(20, 7) NOT NULL DEFAULT 0,
    total_shares        NUMERIC(20, 7) NOT NULL DEFAULT 0,
    total_trustlines    INTEGER NOT NULL DEFAULT 0,
    last_modified_ledger BIGINT NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pool_asset_a ON liquidity_pools (asset_a_code, asset_a_issuer);
CREATE INDEX idx_pool_asset_b ON liquidity_pools (asset_b_code, asset_b_issuer);

CREATE TABLE offers (
    offer_id            BIGINT PRIMARY KEY,
    seller_id           VARCHAR(56) NOT NULL,
    selling_asset_type  SMALLINT NOT NULL,
    selling_asset_code  VARCHAR(12),
    selling_asset_issuer VARCHAR(56),
    buying_asset_type   SMALLINT NOT NULL,
    buying_asset_code   VARCHAR(12),
    buying_asset_issuer VARCHAR(56),
    amount              NUMERIC(20, 7) NOT NULL,
    price_n             BIGINT NOT NULL,
    price_d             BIGINT NOT NULL,
    price               NUMERIC(20, 10) NOT NULL,
    flags               INTEGER NOT NULL DEFAULT 0,
    last_modified_ledger BIGINT NOT NULL,
    sponsor             VARCHAR(56),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_offer_seller ON offers (seller_id);
CREATE INDEX idx_offer_selling ON offers (selling_asset_code, selling_asset_issuer);
CREATE INDEX idx_offer_buying ON offers (buying_asset_code, buying_asset_issuer);
CREATE INDEX idx_offer_pair ON offers (
    selling_asset_code, selling_asset_issuer,
    buying_asset_code, buying_asset_issuer,
    price
);

CREATE TABLE network_stats (
    id                  SERIAL PRIMARY KEY,
    stat_date           DATE NOT NULL UNIQUE,
    total_accounts      BIGINT NOT NULL,
    total_ledgers       BIGINT NOT NULL,
    total_transactions  BIGINT NOT NULL,
    total_operations    BIGINT NOT NULL,
    total_contracts     INTEGER NOT NULL,
    daily_transactions  BIGINT NOT NULL,
    daily_operations    BIGINT NOT NULL,
    daily_new_accounts  INTEGER NOT NULL,
    daily_new_contracts INTEGER NOT NULL,
    avg_ledger_time_ms  INTEGER,
    avg_fee             NUMERIC(20, 7),
    soroban_tx_pct      NUMERIC(5, 2),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stats_date ON network_stats (stat_date DESC);

CREATE TABLE ingestion_state (
    key                 VARCHAR(64) PRIMARY KEY,
    value               TEXT NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
