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

func TestTelegramProvider_Send(t *testing.T) {
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
		testutils.AssertEqual(t, "123456789", payload["chat_id"], "Expected correct chat ID")
		testutils.AssertEqual(t, "Markdown", payload["parse_mode"], "Expected Markdown parse mode")

		// Verify text content
		text := payload["text"].(string)
		testutils.AssertContains(t, text, "✅ Up", "Expected success indicator")
		testutils.AssertContains(t, text, "test-monitor-1", "Expected monitor ID")
		testutils.AssertContains(t, text, "Test Monitor", "Expected monitor name")
		testutils.AssertContains(t, text, "200", "Expected status code")
		testutils.AssertContains(t, text, "100", "Expected latency")

		// Send success response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Setup context with Sentry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())

	// Create Telegram provider
	provider := main.NewTelegramAlertProvider(main.TelegramProviderConfig{
		Url:        server.URL,
		ChatID:     "123456789",
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
	testutils.AssertNoError(t, err, "Failed to send Telegram alert")
}

func TestTelegramProvider_Send_ErrorCases(t *testing.T) {
	// Setup context with Sentry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())

	// Test case: Empty URL
	provider := main.NewTelegramAlertProvider(main.TelegramProviderConfig{
		Url:    "",
		ChatID: "123456789",
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
	testutils.AssertError(t, err, "Expected error for empty URL")

	// Test case: Empty ChatID
	provider = main.NewTelegramAlertProvider(main.TelegramProviderConfig{
		Url:    "https://api.telegram.org/bot123456789:ABC-DEF1234ghIkl-zyx57W2v1u123ew11/sendMessage",
		ChatID: "",
	})

	err = provider.Send(ctx, msg)
	testutils.AssertError(t, err, "Expected error for empty ChatID")

	// Test case: Server returns error status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider = main.NewTelegramAlertProvider(main.TelegramProviderConfig{
		Url:        server.URL,
		ChatID:     "123456789",
		HttpClient: server.Client(),
	})

	err = provider.Send(ctx, msg)
	testutils.AssertError(t, err, "Expected error for server error status")
}

func TestTelegramProvider_Send_Timeout(t *testing.T) {
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

	// Create Telegram provider with short timeout
	client := &http.Client{Timeout: 1 * time.Second}
	provider := main.NewTelegramAlertProvider(main.TelegramProviderConfig{
		Url:        server.URL,
		ChatID:     "123456789",
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
