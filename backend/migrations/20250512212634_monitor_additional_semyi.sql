-- +goose Up
-- +goose StatementBegin
ALTER TABLE monitor_historical ADD COLUMN IF NOT EXISTS additional_message TEXT;
ALTER TABLE monitor_historical ADD COLUMN IF NOT EXISTS http_protocol VARCHAR(255);
ALTER TABLE monitor_historical ADD COLUMN IF NOT EXISTS tls_version VARCHAR(255);
ALTER TABLE monitor_historical ADD COLUMN IF NOT EXISTS tls_cipher VARCHAR(255);
ALTER TABLE monitor_historical ADD COLUMN IF NOT EXISTS tls_expiry TIMESTAMP;

ALTER TABLE monitor_historical_daily_aggregate ADD COLUMN IF NOT EXISTS additional_message TEXT;
ALTER TABLE monitor_historical_daily_aggregate ADD COLUMN IF NOT EXISTS http_protocol VARCHAR(255);
ALTER TABLE monitor_historical_daily_aggregate ADD COLUMN IF NOT EXISTS tls_version VARCHAR(255);
ALTER TABLE monitor_historical_daily_aggregate ADD COLUMN IF NOT EXISTS tls_cipher VARCHAR(255);
ALTER TABLE monitor_historical_daily_aggregate ADD COLUMN IF NOT EXISTS tls_expiry TIMESTAMP;

ALTER TABLE monitor_historical_hourly_aggregate ADD COLUMN IF NOT EXISTS additional_message TEXT;
ALTER TABLE monitor_historical_hourly_aggregate ADD COLUMN IF NOT EXISTS http_protocol VARCHAR(255);
ALTER TABLE monitor_historical_hourly_aggregate ADD COLUMN IF NOT EXISTS tls_version VARCHAR(255);
ALTER TABLE monitor_historical_hourly_aggregate ADD COLUMN IF NOT EXISTS tls_cipher VARCHAR(255);
ALTER TABLE monitor_historical_hourly_aggregate ADD COLUMN IF NOT EXISTS tls_expiry TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE monitor_historical DROP COLUMN IF EXISTS additional_message;
ALTER TABLE monitor_historical DROP COLUMN IF EXISTS http_protocol;
ALTER TABLE monitor_historical DROP COLUMN IF EXISTS tls_version;
ALTER TABLE monitor_historical DROP COLUMN IF EXISTS tls_cipher;
ALTER TABLE monitor_historical DROP COLUMN IF EXISTS tls_expiry;

ALTER TABLE monitor_historical_daily_aggregate DROP COLUMN IF EXISTS additional_message;
ALTER TABLE monitor_historical_daily_aggregate DROP COLUMN IF EXISTS http_protocol;
ALTER TABLE monitor_historical_daily_aggregate DROP COLUMN IF EXISTS tls_version;
ALTER TABLE monitor_historical_daily_aggregate DROP COLUMN IF EXISTS tls_cipher;
ALTER TABLE monitor_historical_daily_aggregate DROP COLUMN IF EXISTS tls_expiry;

ALTER TABLE monitor_historical_hourly_aggregate DROP COLUMN IF EXISTS additional_message;
ALTER TABLE monitor_historical_hourly_aggregate DROP COLUMN IF EXISTS http_protocol;
ALTER TABLE monitor_historical_hourly_aggregate DROP COLUMN IF EXISTS tls_version;
ALTER TABLE monitor_historical_hourly_aggregate DROP COLUMN IF EXISTS tls_cipher;
ALTER TABLE monitor_historical_hourly_aggregate DROP COLUMN IF EXISTS tls_expiry;
-- +goose StatementEnd
