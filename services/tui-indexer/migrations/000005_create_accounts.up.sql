CREATE TABLE accounts (
    id                  VARCHAR(56) PRIMARY KEY,
    sequence            BIGINT NOT NULL,
    balance             NUMERIC(20, 7) NOT NULL,
    buying_liabilities  NUMERIC(20, 7) NOT NULL DEFAULT 0,
    selling_liabilities NUMERIC(20, 7) NOT NULL DEFAULT 0,
    num_subentries      INTEGER NOT NULL DEFAULT 0,
    home_domain         VARCHAR(256),
    flags               INTEGER NOT NULL DEFAULT 0,
    inflation_dest      VARCHAR(56),
    thresholds          JSONB,
    last_modified_ledger BIGINT NOT NULL,
    sponsor             VARCHAR(56),
    num_sponsored       INTEGER NOT NULL DEFAULT 0,
    num_sponsoring      INTEGER NOT NULL DEFAULT 0,
    data_entries        JSONB,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_account_home_domain ON accounts (home_domain)
    WHERE home_domain IS NOT NULL;
CREATE INDEX idx_account_updated ON accounts (updated_at DESC);
CREATE INDEX idx_account_balance ON accounts (balance DESC);

CREATE TABLE trustlines (
    account_id          VARCHAR(56) NOT NULL,
    asset_type          SMALLINT NOT NULL,
    asset_code          VARCHAR(12) NOT NULL,
    asset_issuer        VARCHAR(56) NOT NULL,
    balance             NUMERIC(20, 7) NOT NULL DEFAULT 0,
    limit_amount        NUMERIC(20, 7) NOT NULL,
    buying_liabilities  NUMERIC(20, 7) NOT NULL DEFAULT 0,
    selling_liabilities NUMERIC(20, 7) NOT NULL DEFAULT 0,
    flags               INTEGER NOT NULL DEFAULT 0,
    last_modified_ledger BIGINT NOT NULL,
    sponsor             VARCHAR(56),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (account_id, asset_code, asset_issuer)
);

CREATE INDEX idx_trustline_asset ON trustlines (asset_code, asset_issuer);
CREATE INDEX idx_trustline_balance ON trustlines (asset_code, asset_issuer, balance DESC)
    WHERE balance > 0;

CREATE TABLE account_signers (
    account_id          VARCHAR(56) NOT NULL,
    signer_key          VARCHAR(56) NOT NULL,
    weight              INTEGER NOT NULL,
    type                VARCHAR(32) NOT NULL,
    sponsor             VARCHAR(56),
    last_modified_ledger BIGINT NOT NULL,
    PRIMARY KEY (account_id, signer_key)
);

CREATE INDEX idx_signer_key ON account_signers (signer_key);
