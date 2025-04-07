package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	main "semyi"
	"semyi/testutils"

	"github.com/getsentry/sentry-go"
)

func TestNewServer(t *testing.T) {
	config := main.ServerConfig{
		SSLRedirect:             false,
		Environment:             "test",
		Hostname:                "localhost",
		Port:                    "8080",
		StaticPath:              "/tmp/test-static",
		MonitorHistoricalReader: &main.MonitorHistoricalReader{},
		MonitorHistoricalWriter: &main.MonitorHistoricalWriter{},
		CentralBroker:           &main.Broker[main.MonitorHistorical]{},
		IncidentWriter:          &main.IncidentWriter{},
		MonitorList:             []main.Monitor{},
		ApiKey:                  "test-key",
	}

	server := main.NewServer(config)
	testutils.AssertNotNil(t, server, "Server should not be nil")
	testutils.AssertEqual(t, "localhost:8080", server.Addr, "Server address should match")
}

func TestServer_SnapshotOverview(t *testing.T) {
	t.Skip()
	// Create a test server with mock dependencies
	server := &main.Server{
		HistoricalReader: &main.MonitorHistoricalReader{},
		Monitors: []main.Monitor{
			{UniqueID: "test-1", Name: "Test Monitor 1"},
			{UniqueID: "test-2", Name: "Test Monitor 2"},
		},
	}

	// Create a test request
	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	req := httptest.NewRequestWithContext(ctx, "GET", "/api/overview", nil)
	w := httptest.NewRecorder()

	// Call the handler
	server.SnapshotOverview(w, req)

	// Check response
	testutils.AssertEqual(t, http.StatusOK, w.Code, "HTTP status code should be OK")

	var response struct {
		Monitors []main.Monitor `json:"monitors"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	testutils.AssertNoError(t, err, "Should decode response without error")
	testutils.AssertEqual(t, 2, len(response.Monitors), "Should have 2 monitors")
}

func TestServer_SubmitIncident(t *testing.T) {
	// Create a test server with mock dependencies
	server := &main.Server{
		IncidentWriter: main.NewIncidentWriter(database),
		APIKey:         "test-key",
	}

	// Create test incident data
	incident := struct {
		MonitorID string `json:"monitor_id"`
		Message   string `json:"message"`
	}{
		MonitorID: "test-monitor",
		Message:   "Test incident",
	}

	body, err := json.Marshal(incident)
	testutils.AssertNoError(t, err, "Should marshal incident without error")

	// Create a test request
	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	req := httptest.NewRequestWithContext(ctx, "POST", "/api/incident", bytes.NewBuffer(body))
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	// Call the handler
	server.SubmitIncident(w, req)

	// Check response
	testutils.AssertEqual(t, http.StatusOK, w.Code, "HTTP status code should be OK")
}

func TestServer_SPAHandler(t *testing.T) {
	// Create a temporary directory for static files
	tempDir, err := os.MkdirTemp("", "test-static")
	testutils.AssertNoError(t, err, "Should create temp directory without error")
	defer os.RemoveAll(tempDir)

	// Create a test index.html file
	indexPath := filepath.Join(tempDir, "index.html")
	err = os.WriteFile(indexPath, []byte("<!DOCTYPE html><html><body>Test</body></html>"), 0644)
	testutils.AssertNoError(t, err, "Should write index.html without error")

	// Create a test server
	server := &main.Server{}

	// Create a test request
	ctx := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	req := httptest.NewRequestWithContext(ctx, "GET", "/", nil)
	w := httptest.NewRecorder()

	// Call the handler
	server.SpaHandler(tempDir)(w, req)

	// Check response
	testutils.AssertEqual(t, http.StatusOK, w.Code, "HTTP status code should be OK")
	testutils.AssertContains(t, w.Body.String(), "Test", "Response should contain 'Test'")
}

func TestServer_PushHealthcheck(t *testing.T) {
	t.Skip()
	// Setup context with Sentry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())

	// Create test monitors
	monitors := []main.Monitor{
		{
			UniqueID: "test-monitor-1",
			Name:     "Test Monitor 1",
			Type:     "pull",
		},
		{
			UniqueID: "test-monitor-2",
			Name:     "Test Monitor 2",
			Type:     "push",
		},
	}

	// Create mock alerters
	mockAlerter := &MockAlerter{}

	// Create a real processor with mock dependencies
	processor := &main.Processor{
		TelegramAlertProvider: mockAlerter,
		DiscordAlertProvider:  mockAlerter,
		HTTPAlertProvider:     mockAlerter,
		SlackAlertProvider:    mockAlerter,
		HistoricalWriter:      main.NewMonitorHistoricalWriter(database),
		HistoricalReader:      main.NewMonitorHistoricalReader(database),
		CentralBroker:         main.NewBroker[main.MonitorHistorical](),
	}

	// Create server with test monitors and processor
	server := &main.Server{
		Monitors:         monitors,
		Processor:        processor,
		HistoricalWriter: main.NewMonitorHistoricalWriter(database),
		HistoricalReader: main.NewMonitorHistoricalReader(database),
		CentralBroker:    main.NewBroker[main.MonitorHistorical](),
		IncidentWriter:   main.NewIncidentWriter(database),
		APIKey:           "",
	}

	// Test case: Valid push healthcheck
	t.Run("ValidPushHealthcheck", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/push/test-monitor-1", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		server.PushHealthcheck(w, req)

		testutils.AssertEqual(t, http.StatusOK, w.Code, "Expected status code 200")

		var response main.HttpCommonSuccess
		err := json.NewDecoder(w.Body).Decode(&response)
		testutils.AssertNoError(t, err, "Failed to decode response")
		testutils.AssertEqual(t, "success", response.Message, "Expected success message")
	})

	// Test case: Missing monitor_id
	t.Run("MissingMonitorID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/push/", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		server.PushHealthcheck(w, req)

		testutils.AssertEqual(t, http.StatusBadRequest, w.Code, "Expected status code 400")

		var response main.HttpCommonError
		err := json.NewDecoder(w.Body).Decode(&response)
		testutils.AssertNoError(t, err, "Failed to decode response")
		testutils.AssertEqual(t, "monitor_id is required", response.Error, "Expected error message")
	})

	// Test case: Invalid monitor_id
	t.Run("InvalidMonitorID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/push/invalid-monitor", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		server.PushHealthcheck(w, req)

		testutils.AssertEqual(t, http.StatusBadRequest, w.Code, "Expected status code 400")

		var response main.HttpCommonError
		err := json.NewDecoder(w.Body).Decode(&response)
		testutils.AssertNoError(t, err, "Failed to decode response")
		testutils.AssertEqual(t, "monitor_id is not in the list of monitors", response.Error, "Expected error message")
	})

	// Test case: Push monitor (should fail)
	t.Run("PushMonitor", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/push/test-monitor-2", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		server.PushHealthcheck(w, req)

		testutils.AssertEqual(t, http.StatusBadRequest, w.Code, "Expected status code 400")

		var response main.HttpCommonError
		err := json.NewDecoder(w.Body).Decode(&response)
		testutils.AssertNoError(t, err, "Failed to decode response")
		testutils.AssertEqual(t, "monitor is not a pull monitor", response.Error, "Expected error message")
	})
}
