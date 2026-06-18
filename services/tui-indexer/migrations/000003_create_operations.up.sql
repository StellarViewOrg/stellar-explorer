CREATE TABLE operations (
    id                  BIGSERIAL,
    transaction_id      BIGINT NOT NULL,
    transaction_hash    CHAR(64) NOT NULL,
    application_order   INTEGER NOT NULL,
    type                SMALLINT NOT NULL,
    type_name           VARCHAR(64) NOT NULL,
    source_account      VARCHAR(56),
    asset_code          VARCHAR(12),
    asset_issuer        VARCHAR(56),
    amount              NUMERIC(38, 7),
    destination         VARCHAR(56),
    contract_id         VARCHAR(56),
    function_name       VARCHAR(128),
    details             JSONB NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id, created_at)
);

SELECT create_hypertable('operations', 'created_at',
    chunk_time_interval => INTERVAL '1 week'
);

CREATE INDEX idx_op_tx_hash ON operations (transaction_hash, created_at);
CREATE INDEX idx_op_source ON operations (source_account, created_at DESC);
CREATE INDEX idx_op_type ON operations (type, created_at DESC);
CREATE INDEX idx_op_destination ON operations (destination, created_at DESC)
    WHERE destination IS NOT NULL;
CREATE INDEX idx_op_asset ON operations (asset_code, asset_issuer, created_at DESC)
    WHERE asset_code IS NOT NULL;
CREATE INDEX idx_op_contract ON operations (contract_id, created_at DESC)
    WHERE contract_id IS NOT NULL;
CREATE INDEX idx_op_function ON operations (contract_id, function_name, created_at DESC)
    WHERE function_name IS NOT NULL;

ALTER TABLE operations SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'type',
    timescaledb.compress_orderby = 'created_at DESC'
);
SELECT add_compression_policy('operations', INTERVAL '2 months');
