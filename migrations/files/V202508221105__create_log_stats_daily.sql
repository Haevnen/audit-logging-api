-- flyway: transactional=false

-- Continuous Aggregate for daily log statistics
CREATE MATERIALIZED VIEW IF NOT EXISTS log_stats_daily
WITH (timescaledb.continuous) AS
SELECT
    tenant_id,
    time_bucket('1 day', event_timestamp) AS day,
    action,
    severity,
    COUNT(*) AS log_count
FROM logs
GROUP BY tenant_id, day, action, severity
WITH NO DATA;

-- Add an index to speed up queries by tenant and day
CREATE INDEX IF NOT EXISTS idx_log_stats_daily_tenant_day
    ON log_stats_daily (tenant_id, day);

-- Refresh policy: keep stats for last 90 days (retention period), refresh every 5 min
SELECT add_continuous_aggregate_policy('log_stats_daily',
    start_offset => INTERVAL '90 days',
    end_offset   => INTERVAL '1 hour',
    schedule_interval => INTERVAL '5 minutes');
