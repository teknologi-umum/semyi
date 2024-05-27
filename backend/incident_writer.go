package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type IncidentWriter struct {
	db *sql.DB
}

func NewIncidentWriter(db *sql.DB) *IncidentWriter {
	return &IncidentWriter{
		db: db,
	}
}

func (w *IncidentWriter) Write(ctx context.Context, incident Incident) error {
	conn, err := w.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	timestamp, err := time.Parse(time.RFC3339, incident.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp: %w", err)
	}

	incidentStatus := incident.Status
	if timestamp.After(time.Now()) {
		incidentStatus = IncidentStatusScheduled
	}

	_, err = conn.ExecContext(ctx, "INSERT INTO incident_data (monitor_id, title, description, timestamp, severity, status, created_by) VALUES (?, ?, ?, ?, ?, ?, ?)",
		incident.MonitorID,
		incident.Title,
		incident.Description,
		timestamp,
		incident.Severity,
		incidentStatus,
		incident.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to submit incident: %w", err)
	}

	return nil
}
