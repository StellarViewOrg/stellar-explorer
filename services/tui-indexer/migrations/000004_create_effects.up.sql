CREATE TABLE effects (
    id                  BIGSERIAL,
    operation_id        BIGINT NOT NULL,
    transaction_hash    CHAR(64) NOT NULL,
    type                SMALLINT NOT NULL,
    type_name           VARCHAR(64) NOT NULL,
    account             VARCHAR(56) NOT NULL,
    details             JSONB NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id, created_at)
);

SELECT create_hypertable('effects', 'created_at',
    chunk_time_interval => INTERVAL '1 week'
);

CREATE INDEX idx_effect_account ON effects (account, created_at DESC);
CREATE INDEX idx_effect_type ON effects (type, created_at DESC);
CREATE INDEX idx_effect_op ON effects (operation_id, created_at);
CREATE INDEX idx_effect_tx ON effects (transaction_hash, created_at);

ALTER TABLE effects SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'account',
    timescaledb.compress_orderby = 'created_at DESC'
);
SELECT add_compression_policy('effects', INTERVAL '2 months');
