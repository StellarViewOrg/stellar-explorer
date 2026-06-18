CREATE TABLE contracts (
    contract_id         VARCHAR(56) PRIMARY KEY,
    wasm_hash           CHAR(64),
    creator_account     VARCHAR(56),
    created_ledger      BIGINT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL,
    last_modified_ledger BIGINT NOT NULL,
    contract_type       SMALLINT NOT NULL DEFAULT 0,
    is_sep41_token      BOOLEAN NOT NULL DEFAULT FALSE,
    is_sep50_nft        BOOLEAN NOT NULL DEFAULT FALSE,
    token_name          VARCHAR(256),
    token_symbol        VARCHAR(32),
    token_decimals      INTEGER,
    contract_spec       JSONB,
    storage_entry_count INTEGER NOT NULL DEFAULT 0,
    event_count         BIGINT NOT NULL DEFAULT 0,
    invocation_count    BIGINT NOT NULL DEFAULT 0,
    label               VARCHAR(256),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_contract_creator ON contracts (creator_account);
CREATE INDEX idx_contract_wasm ON contracts (wasm_hash) WHERE wasm_hash IS NOT NULL;
CREATE INDEX idx_contract_type ON contracts (contract_type);
CREATE INDEX idx_contract_token ON contracts (is_sep41_token) WHERE is_sep41_token = TRUE;
CREATE INDEX idx_contract_nft ON contracts (is_sep50_nft) WHERE is_sep50_nft = TRUE;
CREATE INDEX idx_contract_invocations ON contracts (invocation_count DESC);

CREATE TABLE contract_code (
    wasm_hash           CHAR(64) PRIMARY KEY,
    wasm_bytecode       BYTEA NOT NULL,
    wasm_size           INTEGER NOT NULL,
    spec_xdr            TEXT,
    spec_parsed         JSONB,
    verified            BOOLEAN NOT NULL DEFAULT FALSE,
    source_url          VARCHAR(1024),
    source_commit       VARCHAR(64),
    contract_count      INTEGER NOT NULL DEFAULT 1,
    created_ledger      BIGINT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE contract_storage (
    contract_id         VARCHAR(56) NOT NULL,
    key_xdr             TEXT NOT NULL,
    key_decoded         JSONB,
    value_xdr           TEXT NOT NULL,
    value_decoded       JSONB,
    durability          SMALLINT NOT NULL,
    ttl_ledger          BIGINT,
    last_modified_ledger BIGINT NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (contract_id, key_xdr)
);

CREATE INDEX idx_storage_contract ON contract_storage (contract_id);
CREATE INDEX idx_storage_ttl ON contract_storage (ttl_ledger)
    WHERE ttl_ledger IS NOT NULL;
