package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"
)

type MonitorHistoricalWriter struct {
	db *sql.DB
}

func NewMonitorHistoricalWriter(db *sql.DB) *MonitorHistoricalWriter {
	return &MonitorHistoricalWriter{db: db}
}

func (w *MonitorHistoricalWriter) Write(ctx context.Context, historical MonitorHistorical) error {
	span := sentry.StartSpan(ctx, "function", sentry.WithDescription("MonitorHistoricalWriter.Write"))
	span.SetData("semyi.monitor.id", historical.MonitorID)
	span.SetData("semyi.historical.status", historical.Status)
	ctx = span.Context()
	defer span.Finish()

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

	_, err = conn.ExecContext(
		ctx,
		`INSERT INTO
			monitor_historical
			(
				monitor_id,
				status,
				latency,
				timestamp,
				additional_message,
				http_protocol,
				tls_version,
				tls_cipher,
				tls_expiry
			)
		VALUES
			(
				?,
				?,
				?,
				?,
				?,
				?,
				?,
				?,
				?
			)`,
		historical.MonitorID,
		uint8(historical.Status),
		historical.Latency,
		historical.Timestamp,
		sql.NullString{String: historical.AdditionalMessage, Valid: historical.AdditionalMessage != ""},
		sql.NullString{String: historical.HttpProtocol, Valid: historical.HttpProtocol != ""},
		sql.NullString{String: historical.TLSVersion, Valid: historical.TLSVersion != ""},
		sql.NullString{String: historical.TLSCipherName, Valid: historical.TLSCipherName != ""},
		sql.NullTime{Time: historical.TLSExpiryDate, Valid: historical.TLSExpiryDate.IsZero() == false},
	)
	if err != nil {
		return fmt.Errorf("failed to insert historical data: %w", err)
	}

	return nil
}

func (w *MonitorHistoricalWriter) WriteHourly(ctx context.Context, historical MonitorHistorical) error {
	span := sentry.StartSpan(ctx, "function", sentry.WithDescription("MonitorHistoricalWriter.WriteHourly"))
	span.SetData("semyi.monitor.id", historical.MonitorID)
	span.SetData("semyi.historical.status", historical.Status)
	ctx = span.Context()
	defer span.Finish()

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

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM monitor_historical_hourly_aggregate WHERE monitor_id = ? AND timestamp = ?", historical.MonitorID, historical.Timestamp)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Warn().Err(rollbackErr).Msg("failed to rollback transaction")
		}

		return fmt.Errorf("failed to delete hourly historical data: %w", err)
	}

	_, err = conn.ExecContext(
		ctx,
		`INSERT INTO
			monitor_historical_hourly_aggregate
			(
				monitor_id,
				status,
				latency,
				timestamp,
				created_at,
				additional_message,
				http_protocol,
				tls_version,
				tls_cipher,
				tls_expiry
			)
		VALUES
			(
				?,
				?,
				?,
				?,
				?,
				?,
				?,
				?,
				?,
				?
			)`,
		historical.MonitorID,
		uint8(historical.Status),
		historical.Latency,
		historical.Timestamp,
		time.Now().UTC(),
		sql.NullString{String: historical.AdditionalMessage, Valid: historical.AdditionalMessage != ""},
		sql.NullString{String: historical.HttpProtocol, Valid: historical.HttpProtocol != ""},
		sql.NullString{String: historical.TLSVersion, Valid: historical.TLSVersion != ""},
		sql.NullString{String: historical.TLSCipherName, Valid: historical.TLSCipherName != ""},
		sql.NullTime{Time: historical.TLSExpiryDate, Valid: !historical.TLSExpiryDate.IsZero()},
	)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Warn().Err(rollbackErr).Msg("failed to rollback transaction")
		}

		return fmt.Errorf("failed to insert hourly historical data: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (w *MonitorHistoricalWriter) WriteDaily(ctx context.Context, historical MonitorHistorical) error {
	span := sentry.StartSpan(ctx, "function", sentry.WithDescription("MonitorHistoricalWriter.WriteDaily"))
	span.SetData("semyi.monitor.id", historical.MonitorID)
	span.SetData("semyi.historical.status", historical.Status)
	ctx = span.Context()
	defer span.Finish()

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

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM monitor_historical_daily_aggregate WHERE monitor_id = ? AND timestamp = ?", historical.MonitorID, historical.Timestamp)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Warn().Err(rollbackErr).Msg("failed to rollback transaction")
		}

		return fmt.Errorf("failed to delete daily historical data: %w", err)
	}

	_, err = conn.ExecContext(
		ctx,
		`INSERT INTO
			monitor_historical_daily_aggregate
			(
				monitor_id,
				status,
				latency,
				timestamp,
				created_at,
				additional_message,
				http_protocol,
				tls_version,
				tls_cipher,
				tls_expiry
			)
		VALUES
			(
				?,
				?,
				?,
				?,
				?,
				?,
				?,
				?,
				?,
				?
			)`,
		historical.MonitorID,
		uint8(historical.Status),
		historical.Latency,
		historical.Timestamp,
		time.Now().UTC(),
		sql.NullString{String: historical.AdditionalMessage, Valid: historical.AdditionalMessage != ""},
		sql.NullString{String: historical.HttpProtocol, Valid: historical.HttpProtocol != ""},
		sql.NullString{String: historical.TLSVersion, Valid: historical.TLSVersion != ""},
		sql.NullString{String: historical.TLSCipherName, Valid: historical.TLSCipherName != ""},
		sql.NullTime{Time: historical.TLSExpiryDate, Valid: !historical.TLSExpiryDate.IsZero()},
	)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Warn().Err(rollbackErr).Msg("failed to rollback transaction")
		}

		return fmt.Errorf("failed to insert daily historical data: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
