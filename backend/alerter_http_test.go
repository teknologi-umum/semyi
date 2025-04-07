package main_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	main "semyi"
	"semyi/testutils"

	"github.com/getsentry/sentry-go"
)

func TestHTTPProvider_Send(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and content type
		testutils.AssertEqual(t, http.MethodPost, r.Method, "Expected POST method")
		testutils.AssertEqual(t, "application/json", r.Header.Get("Content-Type"), "Expected JSON content type")

		// Parse the request body
		var payload map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		testutils.AssertNoError(t, err, "Failed to decode request body")

		// Verify payload fields
		testutils.AssertEqual(t, true, payload["success"], "Expected success to be true")
		testutils.AssertEqual(t, "test-monitor-1", payload["monitor_id"], "Expected correct monitor ID")
		testutils.AssertEqual(t, "Test Monitor", payload["monitor_name"], "Expected correct monitor name")
		testutils.AssertEqual(t, float64(200), payload["status_code"], "Expected correct status code")
		testutils.AssertEqual(t, float64(100), payload["latency"], "Expected correct latency")

		// Verify timestamp format
		_, err = time.Parse(time.RFC3339, payload["timestamp"].(string))
		testutils.AssertNoError(t, err, "Expected valid RFC3339 timestamp")

		// Send success response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Setup context with Sentry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())

	// Create HTTP provider
	provider := main.NewHTTPAlertProvider(main.HTTPProviderConfig{
		WebhookURL: server.URL,
		HttpClient: server.Client(),
	})

	// Create test alert message
	msg := main.AlertMessage{
		Success:     true,
		StatusCode:  200,
		Timestamp:   time.Now(),
		MonitorID:   "test-monitor-1",
		MonitorName: "Test Monitor",
		Latency:     100,
	}

	// Send the alert
	err := provider.Send(ctx, msg)
	testutils.AssertNoError(t, err, "Failed to send HTTP alert")
}

func TestHTTPProvider_Send_ErrorCases(t *testing.T) {
	// Test case: Empty webhook URL
	provider := main.NewHTTPAlertProvider(main.HTTPProviderConfig{
		WebhookURL: "",
	})

	msg := main.AlertMessage{
		Success:     true,
		StatusCode:  200,
		Timestamp:   time.Now(),
		MonitorID:   "test-monitor-1",
		MonitorName: "Test Monitor",
		Latency:     100,
	}

	err := provider.Send(context.Background(), msg)
	testutils.AssertError(t, err, "Expected error for empty webhook URL")

	// Test case: Invalid webhook URL
	provider = main.NewHTTPAlertProvider(main.HTTPProviderConfig{
		WebhookURL: "invalid-url",
	})

	err = provider.Send(context.Background(), msg)
	testutils.AssertError(t, err, "Expected error for invalid webhook URL")

	// Test case: Server returns error status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider = main.NewHTTPAlertProvider(main.HTTPProviderConfig{
		WebhookURL: server.URL,
		HttpClient: server.Client(),
	})

	err = provider.Send(context.Background(), msg)
	testutils.AssertError(t, err, "Expected error for server error status")
}

func TestHTTPProvider_Send_Timeout(t *testing.T) {
	// Create a slow server that times out
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // Longer than the client timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create HTTP provider with short timeout
	client := &http.Client{Timeout: 1 * time.Second}
	provider := main.NewHTTPAlertProvider(main.HTTPProviderConfig{
		WebhookURL: server.URL,
		HttpClient: client,
	})

	msg := main.AlertMessage{
		Success:     true,
		StatusCode:  200,
		Timestamp:   time.Now(),
		MonitorID:   "test-monitor-1",
		MonitorName: "Test Monitor",
		Latency:     100,
	}

	err := provider.Send(context.Background(), msg)
	testutils.AssertError(t, err, "Expected timeout error")
}
