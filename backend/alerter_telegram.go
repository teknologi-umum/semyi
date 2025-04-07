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

type TelegramProvider struct {
	url        string
	chatID     string
	httpClient *http.Client
}

type TelegramProviderConfig struct {
	Url        string
	ChatID     string
	HttpClient *http.Client
}

func NewTelegramAlertProvider(config TelegramProviderConfig) *TelegramProvider {
	if config.HttpClient == nil {
		config.HttpClient = &http.Client{Timeout: time.Minute}
	}

	return &TelegramProvider{
		url:        config.Url,
		chatID:     config.ChatID,
		httpClient: config.HttpClient,
	}
}

// Ensure TelegramProvider implements Alerter interface
var _ Alerter = (*TelegramProvider)(nil)

func (t *TelegramProvider) Send(ctx context.Context, msg AlertMessage) error {
	if t.url == "" || t.chatID == "" {
		return fmt.Errorf("can't make a telegram alert request: some config is not set")
	}

	span := sentry.StartSpan(ctx, "function", sentry.WithDescription("TelegramProvider.Send"))
	span.SetData("semyi.alert.provider", "telegram")
	span.SetData("semyi.monitor.id", msg.MonitorID)
	ctx = span.Context()
	defer span.Finish()

	// Perhaps we can use a template file instead.
	title := "🔴 Down"
	if msg.Success {
		title = "✅ Up"
	}
	text := fmt.Sprintf(title+`

	**MonitorID:** %s
	**MonitorName:** %s
	**StatusCode:** %d
	**Latency:** %d
	**Timestamp:** %s`,
		msg.MonitorID,
		msg.MonitorName,
		msg.StatusCode,
		msg.Latency,
		msg.Timestamp.Format(time.RFC3339),
	)
	payload := map[string]any{
		"chat_id":    t.chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	payloadByte, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.url, bytes.NewReader(payloadByte))
	if err != nil {
		return fmt.Errorf("failed to send telegram alert: %w", err)
	}
	defer req.Body.Close()

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}
