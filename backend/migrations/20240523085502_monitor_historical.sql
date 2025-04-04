-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS monitor_historical (
    monitor_id VARCHAR(255) NOT NULL,
    status SMALLINT NOT NULL,
    latency INTEGER NOT NULL DEFAULT 0,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (monitor_id, timestamp)
);

CREATE TABLE IF NOT EXISTS monitor_historical_hourly_aggregate (
    timestamp TIMESTAMP NOT NULL,
    monitor_id VARCHAR(255) NOT NULL,
    status SMALLINT NOT NULL,
    latency INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (monitor_id, timestamp)
);

CREATE TABLE IF NOT EXISTS monitor_historical_daily_aggregate (
    timestamp TIMESTAMP NOT NULL,
    monitor_id VARCHAR(255) NOT NULL,
    status SMALLINT NOT NULL,
    latency INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (monitor_id, timestamp)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS monitor_historical;
DROP TABLE IF EXISTS monitor_historical_hourly_aggregate;
DROP TABLE IF EXISTS monitor_historical_daily_aggregate;
-- +goose StatementEnd
