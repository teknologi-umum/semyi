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
