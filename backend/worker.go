package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/rs/zerolog/log"
)

type Response struct {
	Success         bool      `json:"success"`
	StatusCode      int       `json:"statusCode"`
	RequestDuration int64     `json:"requestDuration"`
	Timestamp       time.Time `json:"timestamp"`
	Monitor
}

// Worker should only run checks for a single monitor, with specific type (HTTP or ICMP monitor).
// For each monitor result (success or fail), it should push the result into the monitor processor.
type Worker struct {
	monitor   Monitor
	processor *Processor
}

func NewWorker(monitor Monitor, processor *Processor) (*Worker, error) {
	// Validate the monitor
	_, err := monitor.Validate()
	if err != nil {
		return &Worker{}, err
	}

	// Set default values
	if monitor.Interval == 0 {
		monitor.Interval = DefaultInterval
	}

	if monitor.Timeout == 0 {
		monitor.Timeout = DefaultTimeout
	}

	if monitor.HttpMethod == "" {
		monitor.HttpMethod = http.MethodGet
	}

	if monitor.HttpExpectedStatusCode == "" {
		monitor.HttpExpectedStatusCode = "2xx"
	}

	if monitor.IcmpPacketSize <= 0 {
		monitor.IcmpPacketSize = 56
	}

	return &Worker{
		monitor:   monitor,
		processor: processor,
	}, nil
}

func (w *Worker) Run() {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(w.monitor.Timeout))

		var response Response
		var err error
		// Make the request
		switch w.monitor.Type {
		case MonitorTypeHTTP:
			response, err = w.makeHttpRequest(ctx)
			if err != nil {
				cancel()
				log.Error().Err(err).Msg("failed to make http request")
				continue
			}
			break
		case MonitorTypePing:
			response, err = w.makeIcmpRequest(ctx)
			if err != nil {
				cancel()
				log.Error().Err(err).Msg("failed to make icmp request")
				continue
			}
		}

		// Insert the response to the database
		w.processor.ProcessResponse(response)

		// Sleep for the interval
		time.Sleep(time.Duration(w.monitor.Interval) * time.Second)
	}
}

func (w *Worker) parseExpectedStatusCode(got int) bool {
	// Valid values:
	// * 200 -> Direct 200 status code
	// * 2xx -> Any 2xx status code (200-299)
	// * 200-300 -> Any 200-300 status code (inclusive)
	// * 2xx-3xx -> Any 2xx (200-299) and 3xx (300-399) status code (inclusive)

	if w.monitor.HttpExpectedStatusCode == strconv.Itoa(got) {
		return true
	}

	parts := strings.Split(w.monitor.HttpExpectedStatusCode, "-")
	ok := false
	for _, part := range parts {
		if ok {
			break
		}

		expectedSmallerParts := strings.Split(part, "")
		gotSmallerParts := strings.Split(strconv.Itoa(got), "")

		for i, expectedPart := range expectedSmallerParts {
			if expectedPart == "x" {
				continue
			}

			if expectedPart == gotSmallerParts[i] {
				ok = true
				continue
			}

			if expectedPart != gotSmallerParts[i] {
				ok = false
				break
			}
		}
	}

	return false
}

func (w *Worker) makeHttpRequest(ctx context.Context) (Response, error) {
	timeStart := time.Now().UnixMilli()

	req, err := http.NewRequestWithContext(ctx, w.monitor.HttpMethod, w.monitor.HttpEndpoint, nil)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return Response{}, fmt.Errorf("failed to create request: %w", err)
	}

	if len(w.monitor.HttpHeaders) > 0 {
		for key, value := range w.monitor.HttpHeaders {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{
		Timeout: time.Duration(w.monitor.Timeout) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return Response{}, fmt.Errorf("failed to make request: %w", err)
	}

	timeEnd := time.Now().UnixMilli()
	return Response{
		Success:         w.parseExpectedStatusCode(resp.StatusCode),
		StatusCode:      resp.StatusCode,
		RequestDuration: timeEnd - timeStart,
		Timestamp:       time.Now(),
		Monitor:         w.monitor,
	}, nil
}

func (w *Worker) makeIcmpRequest(ctx context.Context) (Response, error) {
	pinger, err := probing.NewPinger(w.monitor.IcmpHostname)
	if err != nil {
		return Response{}, fmt.Errorf("failed to create pinger: %w", err)
	}

	pinger.Count = 1
	pinger.Size = w.monitor.IcmpPacketSize
	pinger.Timeout = time.Duration(w.monitor.Timeout) * time.Second

	err = pinger.Run()
	if err != nil {
		return Response{}, fmt.Errorf("failed to run pinger: %w", err)
	}

	stats := pinger.Statistics()

	return Response{
		Success:         stats.PacketsRecv > 0,
		StatusCode:      0,
		RequestDuration: int64(stats.AvgRtt / time.Millisecond),
		Timestamp:       time.Now(),
		Monitor:         w.monitor,
	}, nil
}
