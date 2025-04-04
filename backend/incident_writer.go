package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
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
	span := sentry.StartSpan(ctx, "function", sentry.WithDescription("IncidentWriter.Write"))
	span.SetData("semyi.monitor.id", incident.MonitorID)
	span.SetData("semyi.incident.severity", incident.Severity)
	span.SetData("semyi.incident.status", incident.Status)
	ctx = span.Context()
	defer span.Finish()

	conn, err := w.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Ensure timestamps are in UTC
	incident.Timestamp = EnsureUTC(incident.Timestamp)
	now := EnsureUTC(time.Now())

	incidentStatus := incident.Status
	if incident.Timestamp.After(now) {
		incidentStatus = IncidentStatusScheduled
	}

	_, err = conn.ExecContext(ctx, "INSERT INTO incident_data (monitor_id, title, description, timestamp, severity, status, created_by) VALUES (?, ?, ?, ?, ?, ?, ?)",
		incident.MonitorID,
		incident.Title,
		incident.Description,
		incident.Timestamp,
		incident.Severity,
		incidentStatus,
		incident.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to submit incident: %w", err)
	}

	return nil
}
