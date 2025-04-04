package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HTTPProvider struct {
	webhookURL string
	httpClient *http.Client
}

type HTTPProviderConfig struct {
	WebhookURL string
	HttpClient *http.Client
}

func NewHTTPAlertProvider(config HTTPProviderConfig) *HTTPProvider {
	if config.HttpClient == nil {
		config.HttpClient = &http.Client{Timeout: time.Second * 3}
	}

	return &HTTPProvider{
		webhookURL: config.WebhookURL,
		httpClient: config.HttpClient,
	}
}

// Ensure HTTPProvider implements Alerter interface
var _ Alerter = (*HTTPProvider)(nil)

func (h *HTTPProvider) Send(ctx context.Context, msg AlertMessage) error {
	if h.webhookURL == "" {
		return fmt.Errorf("can't make a HTTP webhook request: webhook URL is not set")
	}

	// Format the message as a JSON payload
	payload := map[string]interface{}{
		"success":      msg.Success,
		"monitor_id":   msg.MonitorID,
		"monitor_name": msg.MonitorName,
		"status_code":  msg.StatusCode,
		"latency":      msg.Latency,
		"timestamp":    msg.Timestamp.Format(time.RFC3339),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal HTTP webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.webhookURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create HTTP webhook request: %w", err)
	}
	defer req.Body.Close()

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make HTTP webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("HTTP webhook returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}
