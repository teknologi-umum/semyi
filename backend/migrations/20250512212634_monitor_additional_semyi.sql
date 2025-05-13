-- +goose Up
-- +goose StatementBegin
ALTER TABLE monitor_historical ADD COLUMN additional_message TEXT;
ALTER TABLE monitor_historical ADD COLUMN http_protocol VARCHAR(255);
ALTER TABLE monitor_historical ADD COLUMN tls_version VARCHAR(255);
ALTER TABLE monitor_historical ADD COLUMN tls_cipher VARCHAR(255);
ALTER TABLE monitor_historical ADD COLUMN tls_expiry TIMESTAMP;

ALTER TABLE monitor_historical_daily_aggregate ADD COLUMN additional_message TEXT;
ALTER TABLE monitor_historical_daily_aggregate ADD COLUMN http_protocol VARCHAR(255);
ALTER TABLE monitor_historical_daily_aggregate ADD COLUMN tls_version VARCHAR(255);
ALTER TABLE monitor_historical_daily_aggregate ADD COLUMN tls_cipher VARCHAR(255);
ALTER TABLE monitor_historical_daily_aggregate ADD COLUMN tls_expiry TIMESTAMP;

ALTER TABLE monitor_historical_hourly_aggregate ADD COLUMN additional_message TEXT;
ALTER TABLE monitor_historical_hourly_aggregate ADD COLUMN http_protocol VARCHAR(255);
ALTER TABLE monitor_historical_hourly_aggregate ADD COLUMN tls_version VARCHAR(255);
ALTER TABLE monitor_historical_hourly_aggregate ADD COLUMN tls_cipher VARCHAR(255);
ALTER TABLE monitor_historical_hourly_aggregate ADD COLUMN tls_expiry TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE monitor_historical DROP COLUMN additional_message;
ALTER TABLE monitor_historical DROP COLUMN http_protocol;
ALTER TABLE monitor_historical DROP COLUMN tls_version;
ALTER TABLE monitor_historical DROP COLUMN tls_cipher;
ALTER TABLE monitor_historical DROP COLUMN tls_expiry;

ALTER TABLE monitor_historical_daily_aggregate DROP COLUMN additional_message;
ALTER TABLE monitor_historical_daily_aggregate DROP COLUMN http_protocol;
ALTER TABLE monitor_historical_daily_aggregate DROP COLUMN tls_version;
ALTER TABLE monitor_historical_daily_aggregate DROP COLUMN tls_cipher;
ALTER TABLE monitor_historical_daily_aggregate DROP COLUMN tls_expiry;

ALTER TABLE monitor_historical_hourly_aggregate DROP COLUMN additional_message;
ALTER TABLE monitor_historical_hourly_aggregate DROP COLUMN http_protocol;
ALTER TABLE monitor_historical_hourly_aggregate DROP COLUMN tls_version;
ALTER TABLE monitor_historical_hourly_aggregate DROP COLUMN tls_cipher;
ALTER TABLE monitor_historical_hourly_aggregate DROP COLUMN tls_expiry;
-- +goose StatementEnd
