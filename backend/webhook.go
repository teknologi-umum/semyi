package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Status string

const (
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
)

type WebhookInformation struct {
	Endpoint        string `json:"endpoint"`
	Status          Status `json:"status"`
	StatusCode      int    `json:"statusCode"`
	RequestDuration int64  `json:"requestDuration"`
	Timestamp       int64  `json:"timestamp"`
}

func (w *Webhook) Send(ctx context.Context, response Response) error {
	// Fast return if there is no URL provided
	if w.URL == "" {
		return nil
	}

	var responseStatus Status = StatusFailed
	if response.Success {
		responseStatus = StatusSuccess
	}

	if responseStatus == StatusFailed && !w.FailedResponse {
		return nil
	}

	if responseStatus == StatusSuccess && !w.SuccessResponse {
		return nil
	}

	body, err := json.Marshal(WebhookInformation{
		Endpoint:        response.Endpoint.URL,
		Status:          responseStatus,
		StatusCode:      response.StatusCode,
		RequestDuration: response.RequestDuration,
		Timestamp:       response.Timestamp,
	})
	if err != nil {
		return fmt.Errorf("error marshalling response: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Semyi Webhook")

	client := http.DefaultClient
	_, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}

	return nil
}
