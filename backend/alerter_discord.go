package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
)

type DiscordProvider struct {
	webhookURL string
	httpClient *http.Client
}

type DiscordProviderConfig struct {
	WebhookURL string
	HttpClient *http.Client
}

func NewDiscordAlertProvider(config DiscordProviderConfig) *DiscordProvider {
	if config.HttpClient == nil {
		config.HttpClient = &http.Client{Timeout: time.Minute}
	}
	return &DiscordProvider{
		webhookURL: config.WebhookURL,
		httpClient: config.HttpClient,
	}
}

// Ensure DiscordProvider implements Alerter interface
var _ Alerter = (*DiscordProvider)(nil)

func (d *DiscordProvider) Send(ctx context.Context, msg AlertMessage) error {
	if d.webhookURL == "" {
		return fmt.Errorf("can't make a discord alert request: webhook URL is not set")
	}

	span := sentry.StartSpan(ctx, "function", sentry.WithDescription("DiscordProvider.Send"))
	span.SetData("semyi.alert.provider", "discord")
	span.SetData("semyi.monitor.id", msg.MonitorID)
	ctx = span.Context()
	defer span.Finish()

	// Create a Discord embed message
	title := "🔴 Service Down"
	if msg.Success {
		title = "✅ Service Up"
	}

	// Discord embed color (red for down, green for up)
	color := 0xFF0000 // Red
	if msg.Success {
		color = 0x00FF00 // Green
	}

	// Format the message as a Discord embed
	embed := map[string]interface{}{
		"title": title,
		"color": color,
		"fields": []map[string]string{
			{
				"name":   "Monitor ID",
				"value":  msg.MonitorID,
				"inline": "true",
			},
			{
				"name":   "Monitor Name",
				"value":  msg.MonitorName,
				"inline": "true",
			},
			{
				"name":   "Status Code",
				"value":  fmt.Sprintf("%d", msg.StatusCode),
				"inline": "true",
			},
			{
				"name":   "Latency",
				"value":  fmt.Sprintf("%d ms", msg.Latency),
				"inline": "true",
			},
			{
				"name":   "Timestamp",
				"value":  msg.Timestamp.Format(time.RFC3339),
				"inline": "true",
			},
		},
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{embed},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal discord payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.webhookURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create discord request: %w", err)
	}
	defer req.Body.Close()

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make discord request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discord webhook returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}
