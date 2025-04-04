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
	monitor          Monitor
	processor        *Processor
	historicalReader *MonitorHistoricalReader
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
		var doNotWriteToDatabase = false
		// Make the request
		switch w.monitor.Type {
		case MonitorTypeHTTP:
			log.Debug().Str("monitor_id", w.monitor.UniqueID).Msg("making http request")
			response, err = w.makeHttpRequest(ctx)
			if err != nil {
				cancel()
				log.Error().Err(err).Msg("failed to make http request")
			}
		case MonitorTypePing:
			log.Debug().Str("monitor_id", w.monitor.UniqueID).Msg("making icmp request")
			response, err = w.makeIcmpRequest(ctx)
			if err != nil {
				cancel()
				log.Error().Err(err).Msg("failed to make icmp request")
			}
		case MonitorTypePull:
			log.Debug().Str("monitor_id", w.monitor.UniqueID).Msg("pulling data")
			response, err = w.backfillPullHealthcheck(ctx)
			if err != nil {
				cancel()
				log.Error().Err(err).Msg("failed to make pull request")
			}

			if response.Success {
				doNotWriteToDatabase = true
			}
		}

		if !doNotWriteToDatabase {
			// Insert the response to the database
			log.Debug().Str("monitor_id", w.monitor.UniqueID).Msg("processing response")
			go w.processor.ProcessResponse(response)
		}

		// Sleep for the interval
		log.Debug().Str("monitor_id", w.monitor.UniqueID).Msgf("sleeping for %d seconds", w.monitor.Interval)
		time.Sleep(time.Duration(w.monitor.Interval) * time.Second)
	}
}

func (w *Worker) parseExpectedStatusCode(got int) bool {
	// Valid values:
	// * 200 -> Direct 200 status code
	// * 2xx -> Any 2xx status code (200-299)
	// * 200-300 -> Any 200-300 status code (inclusive)
	// * 2xx-3xx -> Any 2xx (200-299) and 3xx (300-399) status code (inclusive)

	httpExpectedStatusCode := w.monitor.HttpExpectedStatusCode

	if httpExpectedStatusCode == "" {
		httpExpectedStatusCode = "200-399"
	}

	if httpExpectedStatusCode == strconv.Itoa(got) {
		return true
	}

	parts := strings.Split(httpExpectedStatusCode, "-")
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

	return ok
}

func (w *Worker) makeHttpRequest(ctx context.Context) (Response, error) {
	timeStart := time.Now().UnixMilli()

	req, err := http.NewRequestWithContext(ctx, w.monitor.HttpMethod, w.monitor.HttpEndpoint, nil)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return Response{
			Success:         false,
			StatusCode:      0,
			RequestDuration: time.Now().UnixMilli() - timeStart,
			Timestamp:       time.Now().UTC(),
			Monitor:         w.monitor,
		}, fmt.Errorf("failed to create request: %w", err)
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
		return Response{
			Success:         false,
			StatusCode:      0,
			RequestDuration: time.Now().UnixMilli() - timeStart,
			Timestamp:       time.Now().UTC(),
			Monitor:         w.monitor,
		}, fmt.Errorf("failed to make request: %w", err)
	}

	timeEnd := time.Now().UnixMilli()
	return Response{
		Success:         w.parseExpectedStatusCode(resp.StatusCode),
		StatusCode:      resp.StatusCode,
		RequestDuration: timeEnd - timeStart,
		Timestamp:       time.Now().UTC(),
		Monitor:         w.monitor,
	}, nil
}

func (w *Worker) makeIcmpRequest(ctx context.Context) (Response, error) {
	timeStart := time.Now().UnixMilli()

	pinger, err := probing.NewPinger(w.monitor.IcmpHostname)
	if err != nil {
		return Response{
			Success:         false,
			StatusCode:      0,
			RequestDuration: time.Now().UnixMilli() - timeStart,
			Timestamp:       time.Now().UTC(),
			Monitor:         w.monitor,
		}, fmt.Errorf("failed to create pinger: %w", err)
	}

	pinger.Count = 1
	pinger.Size = w.monitor.IcmpPacketSize
	pinger.Timeout = time.Duration(w.monitor.Timeout) * time.Second

	err = pinger.Run()
	if err != nil {
		return Response{
			Success:         false,
			StatusCode:      0,
			RequestDuration: time.Now().UnixMilli() - timeStart,
			Timestamp:       time.Now().UTC(),
			Monitor:         w.monitor,
		}, fmt.Errorf("failed to run pinger: %w", err)
	}

	stats := pinger.Statistics()

	requestDuration := time.Now().UnixMilli() - timeStart
	if stats != nil {
		requestDuration = int64(stats.AvgRtt / time.Millisecond)
	}

	success := false
	if stats != nil {
		success = stats.PacketsRecv > 0
	}

	return Response{
		Success:         success,
		StatusCode:      0,
		RequestDuration: requestDuration,
		Timestamp:       time.Now().UTC(),
		Monitor:         w.monitor,
	}, nil
}

// backfillPullHealthcheck is used to backfill the healthcheck data for pull monitors.
// It will search for the latest historical data for the monitor, if there are no data
// for a certain period, it will set the status to failure.
func (w *Worker) backfillPullHealthcheck(ctx context.Context) (Response, error) {
	historical, err := w.historicalReader.ReadRawLatest(ctx, w.monitor.UniqueID)
	if err != nil {
		return Response{}, fmt.Errorf("failed to read historical data: %w", err)
	}

	// See if `historical.Timestamp` is older than `w.monitor.Interval`
	if historical.Timestamp.Before(time.Now().Add(-time.Duration(w.monitor.Interval) * time.Second)) {
		return Response{
			Success:         false,
			StatusCode:      0,
			RequestDuration: 0,
			Timestamp:       time.Now().UTC(),
			Monitor:         w.monitor,
		}, nil
	}

	return Response{
		Success:         true, // Whatever the actual data is, we consider it as success
		StatusCode:      0,
		RequestDuration: 0,
		Timestamp:       historical.Timestamp,
		Monitor:         w.monitor,
	}, nil
}
