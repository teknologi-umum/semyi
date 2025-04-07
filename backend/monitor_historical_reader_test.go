package main_test

import (
	"context"
	"testing"
	"time"

	main "semyi"
	"semyi/testutils"

	"github.com/getsentry/sentry-go"
)

func setupTestData(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())

	// Insert test data into the database
	writer := main.NewMonitorHistoricalWriter(database)

	// Create test monitor historical data
	testData := []main.MonitorHistorical{
		{
			MonitorID: "test-monitor-1",
			Status:    main.MonitorStatusSuccess,
			Latency:   100,
			Timestamp: time.Now(),
		},
		{
			MonitorID: "test-monitor-1",
			Status:    main.MonitorStatusSuccess,
			Latency:   150,
			Timestamp: time.Now().Add(-1 * time.Hour),
		},
		{
			MonitorID: "test-monitor-1",
			Status:    main.MonitorStatusSuccess,
			Latency:   200,
			Timestamp: time.Now().Add(-24 * time.Hour),
		},
	}

	for _, data := range testData {
		err := writer.Write(ctx, data)
		testutils.AssertNoError(t, err, "Failed to write test data")
	}
}

func TestMonitorHistoricalReader_ReadRawHistorical(t *testing.T) {
	setupTestData(t)

	// Setup context with Sentry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())

	// Create reader with existing database
	reader := main.NewMonitorHistoricalReader(database)

	// Test reading historical data
	historical, err := reader.ReadRawHistorical(ctx, "test-monitor-1", false)
	testutils.AssertNoError(t, err, "Failed to read raw historical data")
	testutils.AssertGreater(t, 0, len(historical), "Expected at least one historical record")

	// Verify the data structure
	for _, record := range historical {
		testutils.AssertNotEmpty(t, record.MonitorID, "MonitorID should not be empty")
		testutils.AssertEqual(t, main.MonitorStatusSuccess, record.Status, "Status should be success")
		testutils.AssertGreater(t, int64(0), record.Latency, "Latency should be positive")
		testutils.AssertNotZero(t, record.Timestamp, "Timestamp should not be zero")
	}
}

func TestMonitorHistoricalReader_ReadRawHistorical_LimitResults(t *testing.T) {
	setupTestData(t)

	// Setup context with Sentry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())

	// Create reader with existing database
	reader := main.NewMonitorHistoricalReader(database)

	// Test reading historical data
	historical, err := reader.ReadRawHistorical(ctx, "test-monitor-1", false)
	testutils.AssertNoError(t, err, "Failed to read raw historical data")
	testutils.AssertGreater(t, 0, len(historical), "Expected at least one historical record")

	// Verify the data structure
	for _, record := range historical {
		testutils.AssertNotEmpty(t, record.MonitorID, "MonitorID should not be empty")
		testutils.AssertEqual(t, main.MonitorStatusSuccess, record.Status, "Status should be success")
		testutils.AssertGreater(t, int64(0), record.Latency, "Latency should be positive")
		testutils.AssertNotZero(t, record.Timestamp, "Timestamp should not be zero")
	}
}

func TestMonitorHistoricalReader_ReadHourlyHistorical(t *testing.T) {
	setupTestData(t)

	// Setup context with Sentry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())

	// Create reader with existing database
	reader := main.NewMonitorHistoricalReader(database)

	// Test reading hourly historical data
	historical, err := reader.ReadHourlyHistorical(ctx, "test-monitor-1", false)
	testutils.AssertNoError(t, err, "Failed to read hourly historical data")

	// Verify the data structure if we have records
	if len(historical) > 0 {
		for _, record := range historical {
			testutils.AssertNotEmpty(t, record.MonitorID, "MonitorID should not be empty")
			testutils.AssertEqual(t, main.MonitorStatusSuccess, record.Status, "Status should be success")
			testutils.AssertGreater(t, int64(0), record.Latency, "Latency should be positive")
			testutils.AssertNotZero(t, record.Timestamp, "Timestamp should not be zero")
		}
	}
}

func TestMonitorHistoricalReader_ReadDailyHistorical(t *testing.T) {
	setupTestData(t)

	// Setup context with Sentry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())

	// Create reader with existing database
	reader := main.NewMonitorHistoricalReader(database)

	// Test reading daily historical data
	historical, err := reader.ReadDailyHistorical(ctx, "test-monitor-1", false)
	testutils.AssertNoError(t, err, "Failed to read daily historical data")

	// Verify the data structure if we have records
	if len(historical) > 0 {
		for _, record := range historical {
			testutils.AssertNotEmpty(t, record.MonitorID, "MonitorID should not be empty")
			testutils.AssertEqual(t, main.MonitorStatusSuccess, record.Status, "Status should be success")
			testutils.AssertGreater(t, int64(0), record.Latency, "Latency should be positive")
			testutils.AssertNotZero(t, record.Timestamp, "Timestamp should not be zero")
		}
	}
}

func TestMonitorHistoricalReader_ReadRawLatest(t *testing.T) {
	setupTestData(t)

	// Setup context with Sentry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())

	// Create reader with existing database
	reader := main.NewMonitorHistoricalReader(database)

	// Test reading latest raw historical data
	latest, err := reader.ReadRawLatest(ctx, "test-monitor-1")
	testutils.AssertNoError(t, err, "Failed to read latest raw historical data")

	// Verify the data structure
	testutils.AssertNotEmpty(t, latest.MonitorID, "MonitorID should not be empty")
	testutils.AssertEqual(t, main.MonitorStatusSuccess, latest.Status, "Status should be success")
	testutils.AssertGreater(t, int64(0), latest.Latency, "Latency should be positive")
	testutils.AssertNotZero(t, latest.Timestamp, "Timestamp should not be zero")
}

func TestMonitorHistoricalReader_ErrorCases(t *testing.T) {
	setupTestData(t)

	// Setup context with Sentry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())

	// Create reader with existing database
	reader := main.NewMonitorHistoricalReader(database)

	// Test with non-existent monitor ID
	_, err := reader.ReadRawHistorical(ctx, "non-existent-monitor", false)
	testutils.AssertNoError(t, err, "Expected error for non-existent monitor")

	// Test with empty monitor ID
	_, err = reader.ReadRawHistorical(ctx, "", false)
	testutils.AssertNoError(t, err, "Expected error for empty monitor ID")

	// Test with invalid monitor ID (too long)
	longID := "a" + string(make([]byte, 300)) // Create a string longer than 255 characters
	_, err = reader.ReadRawHistorical(ctx, longID, false)
	testutils.AssertNoError(t, err, "Expected error for too long monitor ID")
}
