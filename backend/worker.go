package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/allegro/bigcache/v3"
)

type Response struct {
	Success         bool  `json:"success"`
	StatusCode      int   `json:"statusCode"`
	RequestDuration int64 `json:"requestDuration"`
	Timestamp       int64 `json:"timestamp"`
	Endpoint
}

type Worker struct {
	endpoint *Endpoint
	db       *sql.DB
	queue    *Queue
	cache    *bigcache.BigCache
	webhook  *Webhook
}

func (d *Deps) NewWorker(e Endpoint) (*Worker, error) {
	// Validate the endpoint
	var endpoint = &e
	_, err := ValidateEndpoint(*endpoint)
	if err != nil {
		return &Worker{}, err
	}

	if endpoint.Interval == 0 {
		endpoint.Interval = d.DefaultInterval
	}

	if endpoint.Timeout == 0 {
		endpoint.Timeout = d.DefaultTimeout
	}

	if endpoint.Method == "" {
		endpoint.Method = http.MethodGet
	}

	return &Worker{
		endpoint: endpoint,
		db:       d.DB,
		queue:    d.Queue,
		cache:    d.Cache,
		webhook:  d.Webhook,
	}, nil
}

func (w *Worker) Run() {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(w.endpoint.Timeout))

		// Make the request
		response, err := w.makeRequest(ctx)
		if err != nil {
			cancel()
			log.Printf("Failed to make request: %v", err)
			continue
		}

		// Insert the response to the database
		err = w.addToQueue(response)
		if err != nil {
			cancel()
			log.Printf("Failed to insert response to the database: %v", err)
			continue
		}

		err = w.webhook.Send(ctx, *response)
		if err != nil {
			cancel()
			log.Printf("Failed to send webhook: %v", err)
			continue
		}

		// Sleep for the interval
		time.Sleep(time.Duration(w.endpoint.Interval) * time.Second)
	}
}

func (w *Worker) makeRequest(ctx context.Context) (*Response, error) {
	timeStart := time.Now().UnixMilli()

	req, err := http.NewRequestWithContext(ctx, w.endpoint.Method, w.endpoint.URL, nil)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return &Response{}, fmt.Errorf("failed to create request: %v", err)
	}

	if len(w.endpoint.Headers) > 0 {
		for key, value := range w.endpoint.Headers {
			req.Header.Add(key, value)
		}
	}

	client := http.Client{
		Timeout: time.Duration(w.endpoint.Timeout) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return &Response{}, fmt.Errorf("failed to make request: %v", err)
	}

	timeEnd := time.Now().UnixMilli()
	return &Response{
		Success:         resp.StatusCode == 200,
		StatusCode:      resp.StatusCode,
		RequestDuration: timeEnd - timeStart,
		Timestamp:       time.Now().UnixMilli(),
		Endpoint:        *w.endpoint,
	}, nil
}

func (w *Worker) addToQueue(response *Response) error {
	w.queue.Lock()
	w.queue.Items = append(w.queue.Items, *response)
	w.queue.Unlock()

	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	err = w.cache.Set(w.endpoint.URL, data)
	if err != nil {
		return err
	}
	return nil
}
