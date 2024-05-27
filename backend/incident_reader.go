package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

type IncidentDataReader struct {
	db *sql.DB
}

func NewIncidentDataReader(db *sql.DB) *IncidentDataReader {
	return &IncidentDataReader{db: db}
}

func (r *IncidentDataReader) ReadRelatedIncidents(ctx context.Context, incidentTitle string, monitorID string) ([]Incident, error) {
	dbCon, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := dbCon.Close()
		if err != nil {
			log.Warn().Stack().Err(err).Msg("Failed to close connection")
		}
	}()

	rows, err := dbCon.QueryContext(ctx, "SELECT monitor_id, title, description, timestamp, severity, status FROM incident_data WHERE monitor_id = ? AND title = ? ORDER BY created_by DESC", monitorID, incidentTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to read related incidents: %w", err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Warn().Stack().Err(err).Msg("Failed to close rows")
		}
	}()

	var incidents []Incident
	for rows.Next() {
		var incident Incident
		err := rows.Scan(&incident.MonitorID, &incident.Title, &incident.Description, &incident.Timestamp, &incident.Severity, &incident.Status)
		if err != nil {
			return nil, err
		}
		incidents = append(incidents, incident)
	}

	return incidents, nil
}

func (r *IncidentDataReader) ReadIncidentByTimestamp(ctx context.Context, incidentTitle string, monitorID string, timestamp time.Time) (Incident, error) {
	dbCon, err := r.db.Conn(ctx)
	if err != nil {
		return Incident{}, err
	}
	defer func() {
		err := dbCon.Close()
		if err != nil {
			log.Warn().Stack().Err(err).Msg("Failed to close connection")
		}
	}()

	var incidentDetail Incident
	err = dbCon.
		QueryRowContext(ctx, "SELECT monitor_id, title, description, timestamp, severity, status FROM incident_data WHERE monitor_id = ? AND title = ? AND timestamp = ?", monitorID, incidentTitle, timestamp).
		Scan(incidentDetail.MonitorID, incidentDetail.Title, incidentDetail.Description, incidentDetail.Timestamp, incidentDetail.Severity, incidentDetail.Status)
	if err != nil {
		return Incident{}, err
	}

	return incidentDetail, nil
}
