CREATE TABLE trades (
    id                  BIGSERIAL,
    ledger_sequence     BIGINT NOT NULL,
    offer_id            BIGINT,
    base_account        VARCHAR(56) NOT NULL,
    base_asset_type     SMALLINT NOT NULL,
    base_asset_code     VARCHAR(12),
    base_asset_issuer   VARCHAR(56),
    base_amount         NUMERIC(20, 7) NOT NULL,
    counter_account     VARCHAR(56) NOT NULL,
    counter_asset_type  SMALLINT NOT NULL,
    counter_asset_code  VARCHAR(12),
    counter_asset_issuer VARCHAR(56),
    counter_amount      NUMERIC(20, 7) NOT NULL,
    price_n             BIGINT NOT NULL,
    price_d             BIGINT NOT NULL,
    base_is_seller      BOOLEAN NOT NULL,
    trade_type          SMALLINT NOT NULL DEFAULT 0,
    liquidity_pool_id   CHAR(64),
    created_at          TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id, created_at)
);

SELECT create_hypertable('trades', 'created_at',
    chunk_time_interval => INTERVAL '1 week'
);

CREATE INDEX idx_trade_base ON trades (base_account, created_at DESC);
CREATE INDEX idx_trade_counter ON trades (counter_account, created_at DESC);
CREATE INDEX idx_trade_pair ON trades (
    base_asset_code, base_asset_issuer,
    counter_asset_code, counter_asset_issuer,
    created_at DESC
);
CREATE INDEX idx_trade_ledger ON trades (ledger_sequence, created_at);

ALTER TABLE trades SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = '',
    timescaledb.compress_orderby = 'created_at DESC'
);
SELECT add_compression_policy('trades', INTERVAL '2 months');
