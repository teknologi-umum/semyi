package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// CleanupWorker handles the cleanup of old historical data based on retention period
type CleanupWorker struct {
	db              *sql.DB
	retentionPeriod int
}

// NewCleanupWorker creates a new cleanup worker
func NewCleanupWorker(db *sql.DB, retentionPeriod int) *CleanupWorker {
	return &CleanupWorker{
		db:              db,
		retentionPeriod: retentionPeriod,
	}
}

// Run starts the cleanup worker
func (w *CleanupWorker) Run(ctx context.Context) {
	// Run cleanup every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.Cleanup(ctx); err != nil {
				log.Error().Err(err).Msg("Failed to run cleanup")
			}
		}
	}
}

// Cleanup removes historical data older than the retention period
func (w *CleanupWorker) Cleanup(ctx context.Context) error {
	// Calculate the cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -w.retentionPeriod)

	// Get a connection from the pool
	conn, err := w.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Warn().Err(err).Msg("Failed to close connection")
		}
	}()

	// Start a transaction
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Delete old data from raw historical table
	_, err = tx.ExecContext(ctx, "DELETE FROM monitor_historical WHERE timestamp < ?", cutoffDate)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Error().Err(rollbackErr).Msg("Failed to rollback transaction after raw historical data deletion")
		}
		return fmt.Errorf("failed to delete old raw historical data: %w", err)
	}

	// Delete old data from hourly aggregate table
	_, err = tx.ExecContext(ctx, "DELETE FROM monitor_historical_hourly_aggregate WHERE timestamp < ?", cutoffDate)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Error().Err(rollbackErr).Msg("Failed to rollback transaction after hourly historical data deletion")
		}
		return fmt.Errorf("failed to delete old hourly historical data: %w", err)
	}

	// Delete old data from daily aggregate table
	_, err = tx.ExecContext(ctx, "DELETE FROM monitor_historical_daily_aggregate WHERE timestamp < ?", cutoffDate)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Error().Err(rollbackErr).Msg("Failed to rollback transaction after daily historical data deletion")
		}
		return fmt.Errorf("failed to delete old daily historical data: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Error().Err(rollbackErr).Msg("Failed to rollback transaction after commit failure")
		}
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().
		Time("cutoff_date", cutoffDate).
		Int("retention_period_days", w.retentionPeriod).
		Msg("Successfully cleaned up old historical data")

	return nil
}
