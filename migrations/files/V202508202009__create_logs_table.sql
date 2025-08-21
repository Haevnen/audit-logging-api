CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE IF NOT EXISTS logs (
    id UUID DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    session_id TEXT,
    action TEXT NOT NULL,
    resource TEXT,
    resource_id TEXT,
    severity TEXT NOT NULL DEFAULT 'INFO',
    ip_address INET,
    user_agent TEXT,
    message TEXT,
    before JSONB,
    after JSONB,
    metadata JSONB,
    event_timestamp TIMESTAMPTZ NOT NULL, -- when the event happened

    PRIMARY KEY (tenant_id, event_timestamp, id)
);

-- Convert logs into hypertable
SELECT create_hypertable(
    'logs',
    'event_timestamp',
    partitioning_column => 'tenant_id',
    number_partitions => 8,
    chunk_time_interval => INTERVAL '1 day');

-- Enable native compression
ALTER TABLE logs SET (
    timescaledb.compress,
    timescaledb.compress_orderby = 'event_timestamp DESC',
    timescaledb.compress_segmentby = 'tenant_id'
);

-- Policy: compress logs older than 30 days
SELECT add_compression_policy('logs', INTERVAL '30 days');