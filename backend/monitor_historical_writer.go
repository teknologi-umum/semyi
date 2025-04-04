package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

type MonitorHistoricalWriter struct {
	db *sql.DB
}

func NewMonitorHistoricalWriter(db *sql.DB) *MonitorHistoricalWriter {
	return &MonitorHistoricalWriter{db: db}
}

func (w *MonitorHistoricalWriter) Write(ctx context.Context, historical MonitorHistorical) error {
	// Validate the historical data
	valid, err := historical.Validate()
	if err != nil {
		return err
	}
	if !valid {
		return nil
	}

	// Ensure timestamp is in UTC
	historical.Timestamp = EnsureUTC(historical.Timestamp)

	// Insert the historical data into the database
	conn, err := w.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Warn().Err(err).Msg("failed to close connection")
		}
	}()

	_, err = conn.ExecContext(ctx, "INSERT INTO monitor_historical (monitor_id, status, latency, timestamp) VALUES (?, ?, ?, ?)",
		historical.MonitorID, historical.Status, historical.Latency, historical.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to insert historical data: %w", err)
	}

	return nil
}

func (w *MonitorHistoricalWriter) WriteHourly(ctx context.Context, historical MonitorHistorical) error {
	// Validate the historical data
	valid, err := historical.Validate()
	if err != nil {
		return err
	}
	if !valid {
		return nil
	}

	// Insert the historical data into the database
	conn, err := w.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Warn().Err(err).Msg("failed to close connection")
		}
	}()

	_, err = conn.ExecContext(ctx, "INSERT INTO monitor_historical_hourly_aggregate (monitor_id, status, latency, timestamp, created_at) VALUES (?, ?, ?, ?, ?) ON CONFLICT (monitor_id, timestamp) DO UPDATE SET status = ?, latency = ?, created_at = ?",
		historical.MonitorID, historical.Status, historical.Latency, historical.Timestamp, time.Now(), historical.Status, historical.Latency, time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert hourly historical data: %w", err)
	}

	return nil
}

func (w *MonitorHistoricalWriter) WriteDaily(ctx context.Context, historical MonitorHistorical) error {
	// Validate the historical data
	valid, err := historical.Validate()
	if err != nil {
		return err
	}
	if !valid {
		return nil
	}

	// Insert the historical data into the database
	conn, err := w.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Warn().Err(err).Msg("failed to close connection")
		}
	}()

	_, err = conn.ExecContext(ctx, "INSERT INTO monitor_historical_daily_aggregate (monitor_id, status, latency, timestamp, created_at) VALUES (?, ?, ?, ?, ?) ON CONFLICT (monitor_id, timestamp) DO UPDATE SET status = ?, latency = ?, created_at = ?",
		historical.MonitorID, historical.Status, historical.Latency, historical.Timestamp, time.Now(), historical.Status, historical.Latency, time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert daily historical data: %w", err)
	}

	return nil
}
