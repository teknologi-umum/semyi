package main_test

import (
	"context"
	"errors"
	"testing"
	"time"

	main "semyi"
)

func TestNewMonitorHistoricalWriter(t *testing.T) {
	writer := main.NewMonitorHistoricalWriter(database)
	if writer == nil {
		t.Error("expected MonitorHistoricalWriter, got nil")
	}
}

func TestMonitorHistoricalWriter_Write(t *testing.T) {
	if database == nil {
		t.Skip("Database is nil")
		return
	}

	writer := main.NewMonitorHistoricalWriter(database)
	if writer == nil {
		t.Error("expected MonitorHistoricalWriter, got nil")
		return
	}

	t.Run("Should fail on validation error", func(t *testing.T) {
		err := writer.Write(context.Background(), main.MonitorHistorical{
			MonitorID: "",
			Status:    0,
			Latency:   0,
			Timestamp: time.Time{},
		})
		if err == nil {
			t.Error("expected error, got nil")
		}

		if err != nil {
			var validationError *main.ValidationError
			if !errors.As(err, &validationError) {
				t.Errorf("expected ValidationError, got %T", err)
			}
		}
	})

	t.Run("Should successfully write historical data", func(t *testing.T) {
		err := writer.Write(context.Background(), main.MonitorHistorical{
			MonitorID: "test",
			Status:    main.MonitorStatusSuccess,
			Latency:   1,
			Timestamp: time.Now(),
		})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}

func TestMonitorHistoricalWriter_WriteHourly(t *testing.T) {
	if database == nil {
		t.Skip("Database is nil")
		return
	}

	writer := main.NewMonitorHistoricalWriter(database)
	if writer == nil {
		t.Error("expected MonitorHistoricalWriter, got nil")
		return
	}

	t.Run("Should fail on validation error", func(t *testing.T) {
		err := writer.WriteHourly(context.Background(), main.MonitorHistorical{
			MonitorID: "",
			Status:    0,
			Latency:   0,
			Timestamp: time.Time{},
		})
		if err == nil {
			t.Error("expected error, got nil")
		}

		if err != nil {
			var validationError *main.ValidationError
			if !errors.As(err, &validationError) {
				t.Errorf("expected ValidationError, got %T", err)
			}
		}
	})

	t.Run("Should successfully write hourly historical data", func(t *testing.T) {
		err := writer.WriteHourly(context.Background(), main.MonitorHistorical{
			MonitorID: "test",
			Status:    main.MonitorStatusSuccess,
			Latency:   1,
			Timestamp: time.Now(),
		})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("Should use upsert query for duplicate monitor_id and timestamp", func(t *testing.T) {
		timestamp := time.Now()
		monitorId := "upsert-test"

		// Insert the first record
		err := writer.WriteHourly(context.Background(), main.MonitorHistorical{
			MonitorID: monitorId,
			Status:    main.MonitorStatusSuccess,
			Latency:   1,
			Timestamp: timestamp,
		})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		// Insert the second record
		err = writer.WriteHourly(context.Background(), main.MonitorHistorical{
			MonitorID: monitorId,
			Status:    main.MonitorStatusFailure,
			Latency:   2,
			Timestamp: timestamp,
		})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}

func TestMonitorHistoricalWriter_WriteDaily(t *testing.T) {
	if database == nil {
		t.Skip("Database is nil")
		return
	}

	writer := main.NewMonitorHistoricalWriter(database)
	if writer == nil {
		t.Error("expected MonitorHistoricalWriter, got nil")
		return
	}

	t.Run("Should fail on validation error", func(t *testing.T) {
		err := writer.WriteDaily(context.Background(), main.MonitorHistorical{
			MonitorID: "",
			Status:    0,
			Latency:   0,
			Timestamp: time.Time{},
		})
		if err == nil {
			t.Error("expected error, got nil")
		}

		if err != nil {
			var validationError *main.ValidationError
			if !errors.As(err, &validationError) {
				t.Errorf("expected ValidationError, got %T", err)
			}
		}
	})

	t.Run("Should successfully write daily historical data", func(t *testing.T) {
		err := writer.WriteDaily(context.Background(), main.MonitorHistorical{
			MonitorID: "test",
			Status:    main.MonitorStatusSuccess,
			Latency:   1,
			Timestamp: time.Now(),
		})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("Should use upsert query for duplicate monitor_id and timestamp", func(t *testing.T) {
		timestamp := time.Now()
		monitorId := "upsert-test"

		// Insert the first record
		err := writer.WriteDaily(context.Background(), main.MonitorHistorical{
			MonitorID: monitorId,
			Status:    main.MonitorStatusSuccess,
			Latency:   1,
			Timestamp: timestamp,
		})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		// Insert the second record
		err = writer.WriteDaily(context.Background(), main.MonitorHistorical{
			MonitorID: monitorId,
			Status:    main.MonitorStatusFailure,
			Latency:   2,
			Timestamp: timestamp,
		})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}
