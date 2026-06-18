CREATE TABLE ledgers (
    sequence            BIGINT NOT NULL,
    hash                CHAR(64) NOT NULL,
    prev_hash           CHAR(64) NOT NULL,
    closed_at           TIMESTAMPTZ NOT NULL,
    total_coins         BIGINT NOT NULL,
    fee_pool            BIGINT NOT NULL,
    base_fee            INTEGER NOT NULL,
    base_reserve        INTEGER NOT NULL,
    max_tx_set_size     INTEGER NOT NULL,
    protocol_version    INTEGER NOT NULL,
    transaction_count   INTEGER NOT NULL DEFAULT 0,
    operation_count     INTEGER NOT NULL DEFAULT 0,
    successful_tx_count INTEGER NOT NULL DEFAULT 0,
    failed_tx_count     INTEGER NOT NULL DEFAULT 0,
    tx_set_operation_count INTEGER,
    header_xdr          TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (sequence, closed_at)
);

-- TimescaleDB hypertable with monthly chunks
SELECT create_hypertable('ledgers', 'closed_at',
    chunk_time_interval => INTERVAL '1 month',
    if_not_exists => TRUE
);

-- Unique indexes must include the time dimension for hypertables
CREATE UNIQUE INDEX idx_ledger_hash ON ledgers (hash, closed_at);
CREATE INDEX idx_ledger_sequence ON ledgers (sequence);

-- Compression policy: compress chunks older than 3 months
ALTER TABLE ledgers SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = '',
    timescaledb.compress_orderby = 'closed_at DESC'
);
SELECT add_compression_policy('ledgers', INTERVAL '3 months');
