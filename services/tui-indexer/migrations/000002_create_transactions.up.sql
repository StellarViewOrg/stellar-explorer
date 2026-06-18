CREATE TABLE transactions (
    id                  BIGSERIAL,
    hash                CHAR(64) NOT NULL,
    ledger_sequence     BIGINT NOT NULL,
    application_order   INTEGER NOT NULL,
    account             VARCHAR(56) NOT NULL,
    account_muxed       VARCHAR(69),
    account_sequence    BIGINT NOT NULL,
    fee_charged         BIGINT NOT NULL,
    max_fee             BIGINT NOT NULL,
    operation_count     INTEGER NOT NULL,
    memo_type           SMALLINT NOT NULL DEFAULT 0,
    memo_text           TEXT,
    memo_hash           CHAR(64),
    time_bounds_min     TIMESTAMPTZ,
    time_bounds_max     TIMESTAMPTZ,
    status              SMALLINT NOT NULL,
    is_soroban          BOOLEAN NOT NULL DEFAULT FALSE,
    soroban_resources   JSONB,
    envelope_xdr        TEXT NOT NULL,
    result_xdr          TEXT NOT NULL,
    result_meta_xdr     TEXT,
    fee_meta_xdr        TEXT,
    created_at          TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id, created_at)
);

SELECT create_hypertable('transactions', 'created_at',
    chunk_time_interval => INTERVAL '1 week'
);

CREATE UNIQUE INDEX idx_tx_hash ON transactions (hash, created_at);
CREATE INDEX idx_tx_account ON transactions (account, created_at DESC);
CREATE INDEX idx_tx_ledger ON transactions (ledger_sequence, created_at);
CREATE INDEX idx_tx_soroban ON transactions (is_soroban, created_at DESC)
    WHERE is_soroban = TRUE;

ALTER TABLE transactions SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'account',
    timescaledb.compress_orderby = 'created_at DESC'
);
SELECT add_compression_policy('transactions', INTERVAL '2 months');
