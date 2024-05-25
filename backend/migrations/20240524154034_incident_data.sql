-- +goose Up
-- +goose StatementBegin
CREATE TABLE incident_data (
    incident_id VARCHAR(255) NOT NULL,
    monitor_id VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    severity SMALLINT NOT NULL,
    status SMALLINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL
);

CREATE INDEX incident_data_incident_id_idx ON incident_data (monitor_id, incident_id, timestamp);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE incident_data;
-- +goose StatementEnd