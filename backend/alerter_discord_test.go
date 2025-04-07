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

func TestDiscordProvider_Send(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and content type
		testutils.AssertEqual(t, http.MethodPost, r.Method, "Expected POST method")
		testutils.AssertEqual(t, "application/json", r.Header.Get("Content-Type"), "Expected JSON content type")

		// Parse the request body
		var payload map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		testutils.AssertNoError(t, err, "Failed to decode request body")

		// Verify embeds structure
		embeds := payload["embeds"].([]interface{})
		testutils.AssertEqual(t, 1, len(embeds), "Expected 1 embed")

		// Verify embed content
		embed := embeds[0].(map[string]interface{})
		testutils.AssertEqual(t, "✅ Service Up", embed["title"], "Expected success indicator")
		testutils.AssertEqual(t, float64(0x00FF00), embed["color"], "Expected green color for success")

		// Verify fields
		fields := embed["fields"].([]interface{})
		testutils.AssertEqual(t, 5, len(fields), "Expected 5 fields")

		// Verify field values
		fieldMap := make(map[string]interface{})
		for _, f := range fields {
			field := f.(map[string]interface{})
			fieldMap[field["name"].(string)] = field["value"]
		}

		testutils.AssertEqual(t, "test-monitor-1", fieldMap["Monitor ID"], "Expected correct monitor ID")
		testutils.AssertEqual(t, "Test Monitor", fieldMap["Monitor Name"], "Expected correct monitor name")
		testutils.AssertEqual(t, "200", fieldMap["Status Code"], "Expected correct status code")
		testutils.AssertEqual(t, "100 ms", fieldMap["Latency"], "Expected correct latency")

		// Send success response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Setup context with Sentry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())

	// Create Discord provider
	provider := main.NewDiscordAlertProvider(main.DiscordProviderConfig{
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
	testutils.AssertNoError(t, err, "Failed to send Discord alert")
}

func TestDiscordProvider_Send_ErrorCases(t *testing.T) {
	// Setup context with Sentry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())

	// Test case: Empty webhook URL
	provider := main.NewDiscordAlertProvider(main.DiscordProviderConfig{
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

	provider = main.NewDiscordAlertProvider(main.DiscordProviderConfig{
		WebhookURL: server.URL,
		HttpClient: server.Client(),
	})

	err = provider.Send(ctx, msg)
	testutils.AssertError(t, err, "Expected error for server error status")
}

func TestDiscordProvider_Send_Timeout(t *testing.T) {
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

	// Create Discord provider with short timeout
	client := &http.Client{Timeout: 1 * time.Second}
	provider := main.NewDiscordAlertProvider(main.DiscordProviderConfig{
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
