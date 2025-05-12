package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"
)

type MonitorHistoricalReader struct {
	db *sql.DB
}

func NewMonitorHistoricalReader(db *sql.DB) *MonitorHistoricalReader {
	return &MonitorHistoricalReader{db: db}
}

func (r *MonitorHistoricalReader) ReadRawHistorical(ctx context.Context, monitorId string, limitResults bool) ([]MonitorHistorical, error) {
	span := sentry.StartSpan(ctx, "function", sentry.WithDescription("MonitorHistoricalReader.ReadRawHistorical"))
	span.SetData("semyi.monitor.id", monitorId)
	ctx = span.Context()
	defer span.Finish()

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

	query := "SELECT timestamp, monitor_id, status, latency, additional_message, http_protocol, tls_version, tls_cipher, tls_expiry FROM monitor_historical WHERE monitor_id = ? ORDER BY timestamp DESC"
	if limitResults {
		query += " LIMIT 100"
	}

	rows, err := conn.QueryContext(ctx, query, monitorId)
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
		err := rows.Scan(&row.Timestamp, &row.MonitorID, &row.Status, &row.Latency, &row.AdditionalMessage, &row.HttpProtocol, &row.TLSVersion, &row.TLSCipherName, &row.TLSExpiryDate)
		if err != nil {
			return []MonitorHistorical{}, fmt.Errorf("failed to scan row")
		}

		monitorsHistorical = append(monitorsHistorical, row)
	}

	return monitorsHistorical, nil
}

func (r *MonitorHistoricalReader) ReadHourlyHistorical(ctx context.Context, monitorId string, limitResults bool) ([]MonitorHistorical, error) {
	span := sentry.StartSpan(ctx, "function", sentry.WithDescription("MonitorHistoricalReader.ReadHourlyHistorical"))
	span.SetData("semyi.monitor.id", monitorId)
	ctx = span.Context()
	defer span.Finish()

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

	query := "SELECT timestamp, monitor_id, status, latency, additional_message, http_protocol, tls_version, tls_cipher, tls_expiry FROM monitor_historical_hourly_aggregate WHERE monitor_id = ? ORDER BY timestamp DESC"
	if limitResults {
		query += " LIMIT 100"
	}

	rows, err := conn.QueryContext(ctx, query, monitorId)
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
		err := rows.Scan(&row.Timestamp, &row.MonitorID, &row.Status, &row.Latency, &row.AdditionalMessage, &row.HttpProtocol, &row.TLSVersion, &row.TLSCipherName, &row.TLSExpiryDate)
		if err != nil {
			return []MonitorHistorical{}, fmt.Errorf("failed to scan row")
		}

		monitorsHistorical = append(monitorsHistorical, row)
	}

	return monitorsHistorical, nil
}

func (r *MonitorHistoricalReader) ReadDailyHistorical(ctx context.Context, monitorId string, limitResults bool) ([]MonitorHistorical, error) {
	span := sentry.StartSpan(ctx, "function", sentry.WithDescription("ReadDailyHistorical"))
	span.SetData("semyi.monitor.id", monitorId)
	ctx = span.Context()
	defer span.Finish()

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

	query := "SELECT timestamp, monitor_id, status, latency, additional_message, http_protocol, tls_version, tls_cipher, tls_expiry FROM monitor_historical_daily_aggregate WHERE monitor_id = ? ORDER BY timestamp DESC"
	if limitResults {
		query += " LIMIT 100"
	}

	rows, err := conn.QueryContext(ctx, query, monitorId)
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
		err := rows.Scan(&row.Timestamp, &row.MonitorID, &row.Status, &row.Latency, &row.AdditionalMessage, &row.HttpProtocol, &row.TLSVersion, &row.TLSCipherName, &row.TLSExpiryDate)
		if err != nil {
			return []MonitorHistorical{}, fmt.Errorf("failed to scan row")
		}

		monitorsHistorical = append(monitorsHistorical, row)
	}

	return monitorsHistorical, nil
}

func (r *MonitorHistoricalReader) ReadRawLatest(ctx context.Context, monitorId string) (MonitorHistorical, error) {
	span := sentry.StartSpan(ctx, "function", sentry.WithDescription("MonitorHistoricalReader.ReadRawLatest"))
	span.SetData("semyi.monitor.id", monitorId)
	ctx = span.Context()
	defer span.Finish()

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
	err = conn.QueryRowContext(ctx, "SELECT timestamp, monitor_id, status, latency, additional_message, http_protocol, tls_version, tls_cipher, tls_expiry FROM monitor_historical WHERE monitor_id = ? ORDER BY timestamp DESC LIMIT 1", monitorId).Scan(
		&monitorsHistorical.Timestamp,
		&monitorsHistorical.MonitorID,
		&monitorsHistorical.Status,
		&monitorsHistorical.Latency,
		&monitorsHistorical.AdditionalMessage,
		&monitorsHistorical.HttpProtocol,
		&monitorsHistorical.TLSVersion,
		&monitorsHistorical.TLSCipherName,
		&monitorsHistorical.TLSExpiryDate,
	)
	if err != nil {
		return MonitorHistorical{}, fmt.Errorf("failed to read latest raw historical data: %w", err)
	}

	return monitorsHistorical, nil
}
