package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	probing "github.com/prometheus-community/pro-bing"
	"github.com/rs/zerolog/log"
)

type Response struct {
	Success           bool      `json:"success"`
	StatusCode        int       `json:"statusCode"`
	RequestDuration   int64     `json:"requestDuration"`
	Timestamp         time.Time `json:"timestamp"`
	AdditionalMessage string    `json:"additionalMessage,omitempty"`
	HttpProtocol      string    `json:"httpProtocol,omitempty"`
	TLSVersion        string    `json:"tlsVersion,omitempty"`
	TLSCipherName     string    `json:"tlsCipherName,omitempty"`
	TLSExpiryDate     time.Time `json:"tlsExpiryDate,omitempty"`
	TLSIssuer         string    `json:"tlsIssuer,omitempty"`
	Monitor
}

// Worker should only run checks for a single monitor, with specific type (HTTP or ICMP monitor).
// For each monitor result (success or fail), it should push the result into the monitor processor.
type Worker struct {
	monitor                   Monitor
	processor                 *Processor
	historicalReader          *MonitorHistoricalReader
	enableDumpFailureResponse bool
}

func NewWorker(monitor Monitor, processor *Processor, enableDumpFailureResponse bool) (*Worker, error) {
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
		monitor:                   monitor,
		processor:                 processor,
		enableDumpFailureResponse: enableDumpFailureResponse,
	}, nil
}

func (w *Worker) Run() {
	for {
		baseCtx := context.Background()
		ctx := sentry.SetHubOnContext(baseCtx, sentry.CurrentHub().Clone())
		ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(w.monitor.Timeout))

		span := sentry.StartSpan(ctx, "function", sentry.WithDescription("Worker.Run"))
		span.SetData("semyi.monitor.id", w.monitor.UniqueID)
		span.SetData("semyi.monitor.type", w.monitor.Type)
		ctx = span.Context()
		log.Debug().Str("monitor_id", w.monitor.UniqueID).Msg("running worker")

		var response Response
		var err error
		var doNotWriteToDatabase = false

		// Add breadcrumb for monitor check
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "monitor",
			Message:  fmt.Sprintf("Starting %s check for monitor %s", w.monitor.Type, w.monitor.UniqueID),
			Level:    sentry.LevelInfo,
			Data: map[string]interface{}{
				"monitor_id": w.monitor.UniqueID,
				"type":       w.monitor.Type,
				"interval":   w.monitor.Interval,
				"timeout":    w.monitor.Timeout,
			},
		})

		// Make the request
		switch w.monitor.Type {
		case MonitorTypeHTTP:
			log.Debug().Str("monitor_id", w.monitor.UniqueID).Msg("making http request")
			response, err = w.makeHttpRequest(ctx)
			if err != nil {
				cancel()
				log.Error().Err(err).Msg("failed to make http request")
				sentry.GetHubFromContext(ctx).CaptureException(err)
			}
		case MonitorTypePing:
			log.Debug().Str("monitor_id", w.monitor.UniqueID).Msg("making icmp request")
			response, err = w.makeIcmpRequest(ctx)
			if err != nil {
				cancel()
				log.Error().Err(err).Msg("failed to make icmp request")
				sentry.GetHubFromContext(ctx).CaptureException(err)
			}
		case MonitorTypePull:
			log.Debug().Str("monitor_id", w.monitor.UniqueID).Msg("pulling data")
			response, err = w.backfillPullHealthcheck(ctx)
			if err != nil {
				cancel()
				log.Error().Err(err).Msg("failed to make pull request")
				sentry.GetHubFromContext(ctx).CaptureException(err)
			}

			if response.Success {
				doNotWriteToDatabase = true
			}
		}

		if !doNotWriteToDatabase {
			// Insert the response to the database
			log.Debug().Str("monitor_id", w.monitor.UniqueID).Msg("processing response")
			go w.processor.ProcessResponse(context.WithoutCancel(ctx), response)
		}

		// Sleep for the interval
		log.Debug().Str("monitor_id", w.monitor.UniqueID).Msgf("sleeping for %d seconds", w.monitor.Interval)
		span.Finish()
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
	span := sentry.StartSpan(ctx, "function", sentry.WithDescription("Worker.makeHttpRequest"))
	ctx = span.Context()
	defer span.Finish()

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

	certificateAuthorityPool, err := x509.SystemCertPool()
	if err != nil {
		log.Error().Err(err).Msg("failed to get system certificate pool")
		sentry.GetHubFromContext(ctx).CaptureException(err)
		certificateAuthorityPool = x509.NewCertPool()
	}

	client := &http.Client{
		Timeout: time.Duration(w.monitor.Timeout) * time.Second,
		Transport: &http.Transport{
			// Adapted from http.DefaultTransport
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				// TODO: These should be configurable
				Certificates:       nil,
				RootCAs:            certificateAuthorityPool,
				InsecureSkipVerify: true,
			},
		},
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

	expectedStatusCode := w.parseExpectedStatusCode(resp.StatusCode)
	timeEnd := time.Now().UnixMilli()

	if !expectedStatusCode && w.enableDumpFailureResponse {
		dumpRequest, _ := httputil.DumpRequest(req, true)
		dumpResponse, _ := httputil.DumpResponse(resp, true)
		log.Debug().Str("monitor_id", w.monitor.UniqueID).
			Bytes("request", dumpRequest).
			Bytes("response", dumpResponse).
			Msg("dumping failure response")
	}

	var additionalMessage, tlsVersion, tlsCipherName string
	var tlsExpiryDate time.Time
	var tlsIssuer string
	if resp.TLS != nil {
		tlsVersion = tls.VersionName(resp.TLS.Version)
		tlsCipherName = tls.CipherSuiteName(resp.TLS.CipherSuite)

		var tlsNotBeforeDate time.Time
		var tlsCertificateInvalid bool
		var tlsCertificateInvalidMessage string
		if len(resp.TLS.PeerCertificates) > 0 {
			// According to the Go stdlib docs:
			// The first element is the leaf certificate that the connection is verified against.
			firstPeerCertificate := resp.TLS.PeerCertificates[0]
			if firstPeerCertificate != nil {
				tlsIssuer = firstPeerCertificate.Issuer.String()
				tlsExpiryDate = firstPeerCertificate.NotAfter
				tlsNotBeforeDate = firstPeerCertificate.NotBefore

			}

			for _, certificate := range resp.TLS.PeerCertificates {
				verifiedChains, err := certificate.Verify(x509.VerifyOptions{
					Intermediates: certificateAuthorityPool,
				})
				if err == nil {
					tlsCertificateInvalid = false
					continue
				}

				if err != nil {
					tlsCertificateInvalidMessage = err.Error()
					tlsCertificateInvalid = true
				} else if len(verifiedChains) == 0 {
					tlsCertificateInvalidMessage = "TLS certificate is self-signed or not valid"
					tlsCertificateInvalid = true
				}
			}
		}

		if additionalMessage == "" {
			if tlsNotBeforeDate.After(time.Now()) {
				additionalMessage = "TLS certificate is not valid yet"
			} else if tlsExpiryDate.Before(time.Now()) {
				additionalMessage = "TLS certificate is expired"
			} else if tlsCertificateInvalid {
				additionalMessage = tlsCertificateInvalidMessage
			}
		}
	}

	return Response{
		Success:           expectedStatusCode,
		StatusCode:        resp.StatusCode,
		RequestDuration:   timeEnd - timeStart,
		Timestamp:         time.Now().UTC(),
		Monitor:           w.monitor,
		AdditionalMessage: additionalMessage,
		HttpProtocol:      resp.Proto,
		TLSVersion:        tlsVersion,
		TLSCipherName:     tlsCipherName,
		TLSExpiryDate:     tlsExpiryDate,
		TLSIssuer:         tlsIssuer,
	}, nil
}

func (w *Worker) makeIcmpRequest(ctx context.Context) (Response, error) {
	span := sentry.StartSpan(ctx, "function", sentry.WithDescription("Worker.makeIcmpRequest"))
	defer span.Finish()

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
	span := sentry.StartSpan(ctx, "function", sentry.WithDescription("Worker.backfillPullHealthcheck"))
	ctx = span.Context()
	defer span.Finish()

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
