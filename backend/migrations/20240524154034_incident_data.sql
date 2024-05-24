-- +goose Up
-- +goose StatementBegin
CREATE TABLE incident_data (
    instance_id VARCHAR(255) NOT NULL,
    incident_id VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description VARCHAR(255) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    severity SMALLINT NOT NULL,
    status SMALLINT NOT NULL,
    created_by VARCHAR(255) NOT NULL
);

CREATE INDEX incident_data_incident_id_idx ON incident_data (instance_id, incident_id, timestamp);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE incident_data;
-- +goose StatementEnd