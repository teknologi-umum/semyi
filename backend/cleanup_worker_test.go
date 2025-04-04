package main_test

import (
	"context"
	"testing"
	"time"

	main "semyi"
)

func TestCleanupWorker(t *testing.T) {
	// Insert test data
	now := time.Now()
	oldDate := now.AddDate(0, 0, -5)    // 5 days old
	recentDate := now.AddDate(0, 0, -1) // 1 day old

	// Use unique monitor IDs for the test
	monitorID1 := "cleanup_test1"
	monitorID2 := "cleanup_test2"

	// Insert data into raw historical table
	_, err := database.Exec(`
		INSERT INTO monitor_historical (timestamp, monitor_id, status, latency) VALUES
		(?, ?, 1, 100),
		(?, ?, 1, 100),
		(?, ?, 1, 200),
		(?, ?, 1, 200)
	`, oldDate, monitorID1, recentDate, monitorID1, oldDate, monitorID2, recentDate, monitorID2)
	if err != nil {
		t.Fatalf("Failed to insert test data into monitor_historical: %v", err)
	}

	// Insert data into hourly aggregate table
	_, err = database.Exec(`
		INSERT INTO monitor_historical_hourly_aggregate (timestamp, monitor_id, status, latency) VALUES
		(?, ?, 1, 100),
		(?, ?, 1, 100),
		(?, ?, 1, 200),
		(?, ?, 1, 200)
	`, oldDate, monitorID1, recentDate, monitorID1, oldDate, monitorID2, recentDate, monitorID2)
	if err != nil {
		t.Fatalf("Failed to insert test data into monitor_historical_hourly_aggregate: %v", err)
	}

	// Insert data into daily aggregate table
	_, err = database.Exec(`
		INSERT INTO monitor_historical_daily_aggregate (timestamp, monitor_id, status, latency) VALUES
		(?, ?, 1, 100),
		(?, ?, 1, 100),
		(?, ?, 1, 200),
		(?, ?, 1, 200)
	`, oldDate, monitorID1, recentDate, monitorID1, oldDate, monitorID2, recentDate, monitorID2)
	if err != nil {
		t.Fatalf("Failed to insert test data into monitor_historical_daily_aggregate: %v", err)
	}

	// Register cleanup function to remove test data
	t.Cleanup(func() {
		// Clean up monitor_historical
		_, err := database.Exec("DELETE FROM monitor_historical WHERE monitor_id IN (?, ?)", monitorID1, monitorID2)
		if err != nil {
			t.Logf("Warning: failed to clean up monitor_historical: %v", err)
		}

		// Clean up monitor_historical_hourly_aggregate
		_, err = database.Exec("DELETE FROM monitor_historical_hourly_aggregate WHERE monitor_id IN (?, ?)", monitorID1, monitorID2)
		if err != nil {
			t.Logf("Warning: failed to clean up monitor_historical_hourly_aggregate: %v", err)
		}

		// Clean up monitor_historical_daily_aggregate
		_, err = database.Exec("DELETE FROM monitor_historical_daily_aggregate WHERE monitor_id IN (?, ?)", monitorID1, monitorID2)
		if err != nil {
			t.Logf("Warning: failed to clean up monitor_historical_daily_aggregate: %v", err)
		}
	})

	// Create cleanup worker with 3 days retention period
	worker := main.NewCleanupWorker(database, 3)

	// Run cleanup
	err = worker.Cleanup(context.Background())
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify that old data is deleted
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM monitor_historical WHERE timestamp < ? AND monitor_id IN (?, ?)", now.AddDate(0, 0, -3), monitorID1, monitorID2).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query monitor_historical: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 old records in monitor_historical, got %d", count)
	}

	err = database.QueryRow("SELECT COUNT(*) FROM monitor_historical_hourly_aggregate WHERE timestamp < ? AND monitor_id IN (?, ?)", now.AddDate(0, 0, -3), monitorID1, monitorID2).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query monitor_historical_hourly_aggregate: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 old records in monitor_historical_hourly_aggregate, got %d", count)
	}

	err = database.QueryRow("SELECT COUNT(*) FROM monitor_historical_daily_aggregate WHERE timestamp < ? AND monitor_id IN (?, ?)", now.AddDate(0, 0, -3), monitorID1, monitorID2).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query monitor_historical_daily_aggregate: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 old records in monitor_historical_daily_aggregate, got %d", count)
	}

	// Verify that recent data is preserved
	err = database.QueryRow("SELECT COUNT(*) FROM monitor_historical WHERE timestamp >= ? AND monitor_id IN (?, ?)", now.AddDate(0, 0, -3), monitorID1, monitorID2).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query monitor_historical: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 recent records in monitor_historical, got %d", count)
	}

	err = database.QueryRow("SELECT COUNT(*) FROM monitor_historical_hourly_aggregate WHERE timestamp >= ? AND monitor_id IN (?, ?)", now.AddDate(0, 0, -3), monitorID1, monitorID2).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query monitor_historical_hourly_aggregate: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 recent records in monitor_historical_hourly_aggregate, got %d", count)
	}

	err = database.QueryRow("SELECT COUNT(*) FROM monitor_historical_daily_aggregate WHERE timestamp >= ? AND monitor_id IN (?, ?)", now.AddDate(0, 0, -3), monitorID1, monitorID2).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query monitor_historical_daily_aggregate: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 recent records in monitor_historical_daily_aggregate, got %d", count)
	}
}
