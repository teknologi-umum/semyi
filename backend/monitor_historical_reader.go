package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"
)

type MonitorHistoricalReader struct {
	db *sql.DB
}

func NewMonitorHistoricalReader(db *sql.DB) *MonitorHistoricalReader {
	return &MonitorHistoricalReader{db: db}
}

func (r *MonitorHistoricalReader) ReadRawHistorical(ctx context.Context, monitorId string) ([]MonitorHistorical, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return []MonitorHistorical{}, fmt.Errorf("failed to get connection: %w", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Warn().Stack().Err(err).Msg("failed to close connection")
		}
	}()

	rows, err := conn.QueryContext(ctx, "SELECT timestamp, monitor_id, status, latency FROM monitor_historical WHERE monitor_id = ? ORDER BY timestamp DESC", monitorId)
	if err != nil {
		return []MonitorHistorical{}, fmt.Errorf("failed to read raw historical data: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Warn().Stack().Err(err).Msg("failed to close rows")
		}
	}()

	var monitorsHistorical []MonitorHistorical
	for rows.Next() {
		var row MonitorHistorical
		err := rows.Scan(&row.Timestamp, &row.MonitorID, &row.Status, &row.Latency)
		if err != nil {
			return []MonitorHistorical{}, fmt.Errorf("failed to scan row")
		}

		monitorsHistorical = append(monitorsHistorical, row)
	}

	return monitorsHistorical, nil
}

func (r *MonitorHistoricalReader) ReadHourlyHistorical(ctx context.Context, monitorId string) ([]MonitorHistorical, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return []MonitorHistorical{}, fmt.Errorf("failed to get connection: %w", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Warn().Stack().Err(err).Msg("failed to close connection")
		}
	}()

	rows, err := conn.QueryContext(ctx, "SELECT timestamp, monitor_id, status, latency FROM monitor_historical_hourly_aggregate WHERE monitor_id = ? ORDER BY timestamp DESC", monitorId)
	if err != nil {
		return []MonitorHistorical{}, fmt.Errorf("failed to read hourly historical data: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Warn().Stack().Err(err).Msg("failed to close rows")
		}
	}()

	var monitorsHistorical []MonitorHistorical
	for rows.Next() {
		var row MonitorHistorical
		err := rows.Scan(&row.Timestamp, &row.MonitorID, &row.Status, &row.Latency)
		if err != nil {
			return []MonitorHistorical{}, fmt.Errorf("failed to scan row")
		}

		monitorsHistorical = append(monitorsHistorical, row)
	}

	return monitorsHistorical, nil
}

func (r *MonitorHistoricalReader) ReadDailyHistorical(ctx context.Context, monitorId string) ([]MonitorHistorical, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return []MonitorHistorical{}, fmt.Errorf("failed to get connection: %w", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Warn().Stack().Err(err).Msg("failed to close connection")
		}
	}()

	rows, err := conn.QueryContext(ctx, "SELECT timestamp, monitor_id, status, latency FROM monitor_historical_daily_aggregate WHERE monitor_id = ? ORDER BY timestamp DESC", monitorId)
	if err != nil {
		return []MonitorHistorical{}, fmt.Errorf("failed to read daily historical data: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Warn().Stack().Err(err).Msg("failed to close rows")
		}
	}()

	var monitorsHistorical []MonitorHistorical
	for rows.Next() {
		var row MonitorHistorical
		err := rows.Scan(&row.Timestamp, &row.MonitorID, &row.Status, &row.Latency)
		if err != nil {
			return []MonitorHistorical{}, fmt.Errorf("failed to scan row")
		}

		monitorsHistorical = append(monitorsHistorical, row)
	}

	return monitorsHistorical, nil
}

func (r *MonitorHistoricalReader) ReadRawLatest(ctx context.Context, monitorId string) (MonitorHistorical, error) {
	// Get the latest entry from the raw historical table
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return MonitorHistorical{}, fmt.Errorf("failed to get connection: %w", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Warn().Stack().Err(err).Msg("failed to close connection")
		}
	}()

	var monitorsHistorical MonitorHistorical
	err = conn.QueryRowContext(ctx, "SELECT timestamp, monitor_id, status, latency FROM monitor_historical WHERE monitor_id = ? ORDER BY timestamp DESC LIMIT 1", monitorId).Scan(
		&monitorsHistorical.Timestamp,
		&monitorsHistorical.MonitorID,
		&monitorsHistorical.Status,
		&monitorsHistorical.Latency,
	)
	if err != nil {
		return MonitorHistorical{}, fmt.Errorf("failed to read latest raw historical data: %w", err)
	}

	return monitorsHistorical, nil
}
