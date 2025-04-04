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

type SlackProvider struct {
	webhookURL string
	httpClient *http.Client
}

type SlackProviderConfig struct {
	WebhookURL string
	HttpClient *http.Client
}

func NewSlackAlertProvider(config SlackProviderConfig) *SlackProvider {
	if config.HttpClient == nil {
		config.HttpClient = &http.Client{Timeout: time.Second * 3}
	}

	return &SlackProvider{
		webhookURL: config.WebhookURL,
		httpClient: config.HttpClient,
	}
}

// Ensure SlackProvider implements Alerter interface
var _ Alerter = (*SlackProvider)(nil)

func (s *SlackProvider) Send(ctx context.Context, msg AlertMessage) error {
	if s.webhookURL == "" {
		return fmt.Errorf("can't make a Slack webhook request: webhook URL is not set")
	}

	span := sentry.StartSpan(ctx, "function", sentry.WithDescription("SlackProvider.Send"))
	span.SetData("semyi.alert.provider", "slack")
	span.SetData("semyi.monitor.id", msg.MonitorID)
	ctx = span.Context()
	defer span.Finish()

	// Create a Slack message using Block Kit format
	title := "🔴 Service Down"
	if msg.Success {
		title = "✅ Service Up"
	}

	// Create blocks for the Slack message
	blocks := []map[string]interface{}{
		{
			"type": "header",
			"text": map[string]string{
				"type": "plain_text",
				"text": title,
			},
		},
		{
			"type": "section",
			"fields": []map[string]string{
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*Monitor ID*\n%s", msg.MonitorID),
				},
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*Monitor Name*\n%s", msg.MonitorName),
				},
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*Status Code*\n%d", msg.StatusCode),
				},
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*Latency*\n%d ms", msg.Latency),
				},
			},
		},
		{
			"type": "context",
			"elements": []map[string]string{
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("Timestamp: %s", msg.Timestamp.Format(time.RFC3339)),
				},
			},
		},
	}

	// Create the payload
	payload := map[string]interface{}{
		"text":   fmt.Sprintf("%s: %s (%s)", title, msg.MonitorName, msg.MonitorID),
		"blocks": blocks,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create Slack webhook request: %w", err)
	}
	defer req.Body.Close()

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make Slack webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Slack webhook returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}
