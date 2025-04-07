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

func TestSlackProvider_Send(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and content type
		testutils.AssertEqual(t, http.MethodPost, r.Method, "Expected POST method")
		testutils.AssertEqual(t, "application/json", r.Header.Get("Content-Type"), "Expected JSON content type")

		// Parse the request body
		var payload map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		testutils.AssertNoError(t, err, "Failed to decode request body")

		// Verify text field
		testutils.AssertContains(t, payload["text"].(string), "✅ Service Up", "Expected success indicator")
		testutils.AssertContains(t, payload["text"].(string), "Test Monitor", "Expected monitor name")
		testutils.AssertContains(t, payload["text"].(string), "test-monitor-1", "Expected monitor ID")

		// Verify blocks structure
		blocks := payload["blocks"].([]interface{})
		testutils.AssertEqual(t, 3, len(blocks), "Expected 3 blocks")

		// Verify header block
		headerBlock := blocks[0].(map[string]interface{})
		testutils.AssertEqual(t, "header", headerBlock["type"], "Expected header block type")
		testutils.AssertEqual(t, "✅ Service Up", headerBlock["text"].(map[string]interface{})["text"], "Expected success indicator")

		// Verify fields block
		fieldsBlock := blocks[1].(map[string]interface{})
		testutils.AssertEqual(t, "section", fieldsBlock["type"], "Expected section block type")
		fields := fieldsBlock["fields"].([]interface{})
		testutils.AssertEqual(t, 4, len(fields), "Expected 4 fields")

		// Send success response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Setup context with Sentry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())

	// Create Slack provider
	provider := main.NewSlackAlertProvider(main.SlackProviderConfig{
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
	testutils.AssertNoError(t, err, "Failed to send Slack alert")
}

func TestSlackProvider_Send_ErrorCases(t *testing.T) {
	// Setup context with Sentry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())

	// Test case: Empty webhook URL
	provider := main.NewSlackAlertProvider(main.SlackProviderConfig{
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

	err := provider.Send(ctx, msg)
	testutils.AssertError(t, err, "Expected error for empty webhook URL")

	// Test case: Server returns error status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider = main.NewSlackAlertProvider(main.SlackProviderConfig{
		WebhookURL: server.URL,
		HttpClient: server.Client(),
	})

	err = provider.Send(ctx, msg)
	testutils.AssertError(t, err, "Expected error for server error status")
}

func TestSlackProvider_Send_Timeout(t *testing.T) {
	// Setup context with Sentry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())

	// Create a slow server that times out
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // Longer than the client timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create Slack provider with short timeout
	client := &http.Client{Timeout: 1 * time.Second}
	provider := main.NewSlackAlertProvider(main.SlackProviderConfig{
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

	err := provider.Send(ctx, msg)
	testutils.AssertError(t, err, "Expected timeout error")
}
