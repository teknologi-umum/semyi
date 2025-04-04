-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS incident_data (
    monitor_id VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    severity SMALLINT NOT NULL,
    status SMALLINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    PRIMARY KEY (monitor_id, timestamp)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS incident_data;
-- +goose StatementEnd