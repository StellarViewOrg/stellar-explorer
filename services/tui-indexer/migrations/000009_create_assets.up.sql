CREATE TABLE assets (
    id                  SERIAL PRIMARY KEY,
    asset_type          SMALLINT NOT NULL,
    asset_code          VARCHAR(12) NOT NULL,
    asset_issuer        VARCHAR(56) NOT NULL,
    num_accounts        INTEGER NOT NULL DEFAULT 0,
    total_supply        NUMERIC(38, 7) NOT NULL DEFAULT 0,
    num_claimable_balances INTEGER NOT NULL DEFAULT 0,
    num_liquidity_pools INTEGER NOT NULL DEFAULT 0,
    num_contracts       INTEGER NOT NULL DEFAULT 0,
    flags               INTEGER NOT NULL DEFAULT 0,
    auth_required       BOOLEAN NOT NULL DEFAULT FALSE,
    auth_revocable      BOOLEAN NOT NULL DEFAULT FALSE,
    auth_immutable      BOOLEAN NOT NULL DEFAULT FALSE,
    clawback_enabled    BOOLEAN NOT NULL DEFAULT FALSE,
    home_domain         VARCHAR(256),
    toml_name           VARCHAR(256),
    toml_description    TEXT,
    toml_image_url      VARCHAR(1024),
    toml_conditions     TEXT,
    toml_metadata       JSONB,
    sac_contract_id     VARCHAR(56),
    last_trade_at       TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (asset_code, asset_issuer)
);

CREATE INDEX idx_asset_code ON assets (asset_code);
CREATE INDEX idx_asset_issuer ON assets (asset_issuer);
CREATE INDEX idx_asset_accounts ON assets (num_accounts DESC);
CREATE INDEX idx_asset_domain ON assets (home_domain) WHERE home_domain IS NOT NULL;
CREATE INDEX idx_asset_sac ON assets (sac_contract_id) WHERE sac_contract_id IS NOT NULL;
