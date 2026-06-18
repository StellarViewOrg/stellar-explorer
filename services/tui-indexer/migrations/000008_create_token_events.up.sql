CREATE TABLE token_events (
    id                  BIGSERIAL,
    event_type          SMALLINT NOT NULL,
    event_type_name     VARCHAR(16) NOT NULL,
    from_address        VARCHAR(56),
    from_muxed          VARCHAR(69),
    to_address          VARCHAR(56),
    to_muxed            VARCHAR(69),
    to_muxed_id         BIGINT,
    asset_type          SMALLINT NOT NULL,
    asset_code          VARCHAR(12),
    asset_issuer        VARCHAR(56),
    asset_contract_id   VARCHAR(56),
    amount              NUMERIC(38, 0) NOT NULL,
    amount_formatted    NUMERIC(38, 7),
    transaction_hash    CHAR(64) NOT NULL,
    ledger_sequence     BIGINT NOT NULL,
    operation_index     INTEGER,
    created_at          TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id, created_at)
);

SELECT create_hypertable('token_events', 'created_at',
    chunk_time_interval => INTERVAL '1 week'
);

CREATE INDEX idx_token_from ON token_events (from_address, created_at DESC);
CREATE INDEX idx_token_to ON token_events (to_address, created_at DESC);
CREATE INDEX idx_token_asset ON token_events (asset_code, asset_issuer, created_at DESC);
CREATE INDEX idx_token_contract_asset ON token_events (asset_contract_id, created_at DESC)
    WHERE asset_contract_id IS NOT NULL;
CREATE INDEX idx_token_type_asset ON token_events (event_type, asset_code, asset_issuer, created_at DESC);
CREATE INDEX idx_token_tx ON token_events (transaction_hash, created_at);
CREATE INDEX idx_token_ledger ON token_events (ledger_sequence, created_at);

ALTER TABLE token_events SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'event_type',
    timescaledb.compress_orderby = 'created_at DESC'
);
SELECT add_compression_policy('token_events', INTERVAL '2 months');
