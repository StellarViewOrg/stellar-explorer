CREATE TABLE contract_events (
    id                  BIGSERIAL,
    contract_id         VARCHAR(56) NOT NULL,
    transaction_hash    CHAR(64) NOT NULL,
    ledger_sequence     BIGINT NOT NULL,
    type                SMALLINT NOT NULL,
    topic_1             TEXT,
    topic_2             TEXT,
    topic_3             TEXT,
    topic_4             TEXT,
    topics_xdr          TEXT NOT NULL,
    value_xdr           TEXT NOT NULL,
    topics_decoded      JSONB,
    value_decoded       JSONB,
    created_at          TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id, created_at)
);

SELECT create_hypertable('contract_events', 'created_at',
    chunk_time_interval => INTERVAL '1 week'
);

CREATE INDEX idx_event_contract ON contract_events (contract_id, created_at DESC);
CREATE INDEX idx_event_topic1 ON contract_events (topic_1, created_at DESC);
CREATE INDEX idx_event_contract_topic ON contract_events (contract_id, topic_1, created_at DESC);
CREATE INDEX idx_event_tx ON contract_events (transaction_hash, created_at);
CREATE INDEX idx_event_ledger ON contract_events (ledger_sequence, created_at);

ALTER TABLE contract_events SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'contract_id',
    timescaledb.compress_orderby = 'created_at DESC'
);
SELECT add_compression_policy('contract_events', INTERVAL '2 months');
