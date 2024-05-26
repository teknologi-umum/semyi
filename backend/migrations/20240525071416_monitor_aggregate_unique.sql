-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS monitor_historical_hourly_aggregate_monitor_id_timestamp_idx;
DROP INDEX IF EXISTS monitor_historical_daily_aggregate_monitor_id_timestamp_idx;

CREATE UNIQUE INDEX IF NOT EXISTS monitor_historical_hourly_aggregate_monitor_id_timestamp_idx ON monitor_historical_hourly_aggregate (monitor_id, timestamp);
CREATE UNIQUE INDEX IF NOT EXISTS monitor_historical_daily_aggregate_monitor_id_timestamp_idx ON monitor_historical_daily_aggregate (monitor_id, timestamp);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS monitor_historical_hourly_aggregate_monitor_id_timestamp_idx;
DROP INDEX IF EXISTS monitor_historical_daily_aggregate_monitor_id_timestamp_idx;

CREATE INDEX IF NOT EXISTS monitor_historical_hourly_aggregate_monitor_id_timestamp_idx ON monitor_historical_hourly_aggregate (monitor_id, timestamp);
CREATE INDEX IF NOT EXISTS monitor_historical_daily_aggregate_monitor_id_timestamp_idx ON monitor_historical_daily_aggregate (monitor_id, timestamp);
-- +goose StatementEnd
