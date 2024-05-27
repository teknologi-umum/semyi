package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Alerter interface {
	Send(ctx context.Context, msg AlertMessage) error
}

type AlertMessage struct {
	Success     bool
	StatusCode  int
	Timestamp   time.Time
	MonitorID   string
	MonitorName string
	Latency     int64
}

type TelegramProvider struct {
	url    string
	chatID string
}

type TelegramProviderConfig struct {
	Url    string
	ChatID string
}

func NewTelegramAlertProvider(config TelegramProviderConfig) *TelegramProvider {
	return &TelegramProvider{
		url: config.Url,
	}
}

func (t TelegramProvider) Send(ctx context.Context, msg AlertMessage) error {
	if t.url == "" || t.chatID == "" {
		return fmt.Errorf("can't make a telegram alert request: some config is not set")
	}

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

	client := http.Client{Timeout: time.Second * 3}

	_, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	return nil
}
