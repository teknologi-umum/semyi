package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"github.com/unrolled/secure"
)

type Server struct {
	historicalWriter *MonitorHistoricalWriter
	historicalReader *MonitorHistoricalReader
	centralBroker    *Broker[MonitorHistorical]
	incidentWriter   *IncidentWriter
	monitors         []Monitor
	processor        *Processor

	apiKey string
}

type ServerConfig struct {
	SSLRedirect             bool
	Environment             string
	Hostname                string
	Port                    string
	StaticPath              string
	MonitorHistoricalReader *MonitorHistoricalReader
	MonitorHistoricalWriter *MonitorHistoricalWriter
	CentralBroker           *Broker[MonitorHistorical]
	IncidentWriter          *IncidentWriter
	MonitorList             []Monitor

	ApiKey string
}

func NewServer(config ServerConfig) *http.Server {
	server := &Server{
		historicalReader: config.MonitorHistoricalReader,
		historicalWriter: config.MonitorHistoricalWriter,
		centralBroker:    config.CentralBroker,
		monitors:         config.MonitorList,
		incidentWriter:   config.IncidentWriter,
		processor:        nil,

		apiKey: config.ApiKey,
	}

	secureMiddleware := secure.New(secure.Options{
		BrowserXssFilter:   true,
		ContentTypeNosniff: true,
		SSLRedirect:        config.SSLRedirect,
		IsDevelopment:      config.Environment == "development",
	})

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	})

	api := chi.NewRouter()
	api.Use(corsMiddleware.Handler)
	api.Use(middleware.RequestID)
	api.Get("/api/overview", server.snapshotOverview)
	api.Get("/api/by", server.snapshotBy)
	api.Get("/api/static", server.staticSnapshot)
	api.Post("/api/incident", server.submitIncindent)
	api.Get("/api/push/{monitor_id}", server.pushHealthcheck)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(secureMiddleware.Handler)
	r.Handle("/api/*", corsMiddleware.Handler(api))
	r.Handle("/", http.FileServer(http.Dir(config.StaticPath)))

	return &http.Server{
		Addr:    net.JoinHostPort(config.Hostname, config.Port),
		Handler: r,
	}
}

func (s *Server) snapshotOverview(w http.ResponseWriter, r *http.Request) {
	requestId := middleware.GetReqID(r.Context())
	ctx := r.Context()

	// Add breadcrumb for request
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "http",
		Message:  "Handling snapshot overview request",
		Level:    sentry.LevelInfo,
		Data: map[string]interface{}{
			"request_id": requestId,
			"path":       r.URL.Path,
		},
	})

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPreconditionFailed)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "not flusher"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	log.Debug().Str("request_id", requestId).Str("component", "snapshotOverview").Msg("snapshot overview server-sent-event requested, trying to subscribe to endpoints")
	subscriber, err := NewSubscriber(s.centralBroker, monitorIds...)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: fmt.Sprintf("failed to subscribe to endpoints: %s", err)})
		sentry.GetHubFromContext(ctx).CaptureException(err)
		return
	}

	_, err = w.Write([]byte("data: {\"type\": \"hello\"}\n\n"))
	if err != nil {
		log.Error().Str("request_id", requestId).Str("component", "snapshotOverview").Err(err).Msg("failed to write data")
		sentry.GetHubFromContext(ctx).CaptureException(err)
	}
	flusher.Flush()

	for {
		select {
		case <-ctx.Done():
			log.Debug().Str("request_id", requestId).Str("component", "snapshotOverview").Msg("context done, closing")
			return
		case data := <-subscriber.Listen(ctx):
			log.Debug().Str("request_id", requestId).Str("component", "snapshotOverview").Msg("received data from endpoint")
			marshaled, err := json.Marshal(data)
			if err != nil {
				log.Error().Str("request_id", requestId).Str("component", "snapshotOverview").Err(err).Msg("failed to marshal data")
				sentry.GetHubFromContext(ctx).CaptureException(err)
			}

			_, err = w.Write([]byte("data: " + string(marshaled) + "\n\n"))
			if err != nil {
				log.Error().Str("request_id", requestId).Str("component", "snapshotOverview").Err(err).Msg("failed to write data")
				sentry.GetHubFromContext(ctx).CaptureException(err)
			}

			flusher.Flush()
		default:
			time.Sleep(time.Second)
		}
	}
}

func (s *Server) snapshotBy(w http.ResponseWriter, r *http.Request) {
	requestId := middleware.GetReqID(r.Context())
	ctx := r.Context()

	// Add breadcrumb for request
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "http",
		Message:  "Handling snapshot by request",
		Level:    sentry.LevelInfo,
		Data: map[string]interface{}{
			"request_id": requestId,
			"path":       r.URL.Path,
		},
	})

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPreconditionFailed)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "not flusher"})
		return
	}

	ids := r.URL.Query().Get("ids")
	if ids == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "ids is required"})
		return
	}

	wantedMonitorIds := strings.Split(ids, ",")

	for _, id := range wantedMonitorIds {
		if !slices.Contains(monitorIds, id) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "id is not in the list of monitors"})
			return
		}
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	log.Debug().Str("wanted_monitor_ids", ids).Str("request_id", requestId).Str("component", "snapshotBy").Msg("snapshot by server-sent-event requested, trying to subscribe to endpoints")
	sub, err := NewSubscriber(s.centralBroker, wantedMonitorIds...)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: fmt.Sprintf("failed to subscribe to endpoints: %s", err)})
		sentry.GetHubFromContext(ctx).CaptureException(err)
		return
	}

	_, err = w.Write([]byte("data: {\"type\": \"hello\"}\n\n"))
	if err != nil {
		log.Error().Str("wanted_monitor_ids", ids).Str("request_id", requestId).Str("component", "snapshotBy").Err(err).Msg("failed to write data")
		sentry.GetHubFromContext(ctx).CaptureException(err)
	}
	flusher.Flush()

	for {
		select {
		case <-ctx.Done():
			log.Debug().Str("wanted_monitor_ids", ids).Str("request_id", requestId).Str("component", "snapshotBy").Msg("context done, closing")
			return
		case data := <-sub.Listen(ctx):
			log.Debug().Str("wanted_monitor_ids", ids).Str("request_id", requestId).Str("component", "snapshotBy").Msg("received data from endpoint")
			marshaled, err := json.Marshal(data)
			if err != nil {
				log.Error().Str("wanted_monitor_ids", ids).Str("request_id", requestId).Str("component", "snapshotBy").Err(err).Msg("failed to marshal data")
				sentry.GetHubFromContext(ctx).CaptureException(err)
			}

			_, err = w.Write([]byte("data: " + string(marshaled) + "\n\n"))
			if err != nil {
				log.Error().Str("wanted_monitor_ids", ids).Str("request_id", requestId).Str("component", "snapshotBy").Err(err).Msg("failed to write data")
				sentry.GetHubFromContext(ctx).CaptureException(err)
			}

			flusher.Flush()
		default:
			time.Sleep(time.Second)
		}
	}
}

func (s *Server) staticSnapshot(w http.ResponseWriter, r *http.Request) {
	monitorId := r.URL.Query().Get("id")
	ctx := r.Context()

	// Add breadcrumb for request
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "http",
		Message:  "Handling static snapshot request",
		Level:    sentry.LevelInfo,
		Data: map[string]interface{}{
			"monitor_id": monitorId,
			"path":       r.URL.Path,
		},
	})

	if monitorId == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "id is required"})
		return
	}

	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "hourly"
	}

	if interval != "hourly" && interval != "daily" && interval != "raw" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "interval must be hourly, daily, or raw"})
		return
	}

	if !slices.Contains(monitorIds, monitorId) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "id is not in the list of monitors"})
		return
	}

	var err error
	var monitor Monitor
	var monitorHistorical []MonitorHistorical
	switch interval {
	case "raw":
		monitorHistorical, err = s.historicalReader.ReadRawHistorical(ctx, monitorId)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(HttpCommonError{Error: fmt.Sprintf("failed to read raw historical data: %s", err)})
			sentry.GetHubFromContext(ctx).CaptureException(err)
			return
		}
	case "hourly":
		monitorHistorical, err = s.historicalReader.ReadHourlyHistorical(ctx, monitorId)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(HttpCommonError{Error: fmt.Sprintf("failed to read hourly historical data: %s", err)})
			sentry.GetHubFromContext(ctx).CaptureException(err)
			return
		}
	case "daily":
		monitorHistorical, err = s.historicalReader.ReadDailyHistorical(ctx, monitorId)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(HttpCommonError{Error: fmt.Sprintf("failed to read daily historical data: %s", err)})
			sentry.GetHubFromContext(ctx).CaptureException(err)
			return
		}
	}

	// Acquire monitor metadata
	for _, m := range s.monitors {
		if m.UniqueID == monitorId {
			monitor = m
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(StaticSnapshotResponse{
		Metadata:   monitor,
		Historical: monitorHistorical,
	})
}

func (s *Server) submitIncindent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Add breadcrumb for request
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "http",
		Message:  "Handling incident submission",
		Level:    sentry.LevelInfo,
		Data: map[string]interface{}{
			"path": r.URL.Path,
		},
	})

	if s.apiKey != "" {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "X-API-Key is required"})
			return
		}

		if apiKey != s.apiKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "invalid X-API-Key"})
			return
		}
	}

	var incident Incident
	err := json.NewDecoder(r.Body).Decode(&incident)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: fmt.Sprintf("failed to decode request body: %s", err)})
		sentry.GetHubFromContext(ctx).CaptureException(err)
		return
	}

	err = s.incidentWriter.Write(ctx, incident)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: fmt.Sprintf("failed to write incident: %s", err)})
		sentry.GetHubFromContext(ctx).CaptureException(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HttpCommonSuccess{Message: "success"})
}

func (s *Server) pushHealthcheck(w http.ResponseWriter, r *http.Request) {
	monitorId := chi.URLParam(r, "monitor_id")

	// Add breadcrumb for request
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "http",
		Message:  "Handling push healthcheck",
		Level:    sentry.LevelInfo,
		Data: map[string]interface{}{
			"monitor_id": monitorId,
			"path":       r.URL.Path,
		},
	})

	if monitorId == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "monitor_id is required"})
		return
	}

	if !slices.Contains(monitorIds, monitorId) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "monitor_id is not in the list of monitors"})
		return
	}

	var monitor Monitor
	for _, m := range s.monitors {
		if m.UniqueID == monitorId {
			monitor = m
			break
		}
	}

	if monitor.UniqueID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "monitor not found"})
		return
	}

	if monitor.Type != MonitorTypePull {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "monitor is not a pull monitor"})
		return
	}

	response := Response{
		Success:         true,
		StatusCode:      200,
		RequestDuration: 0,
		Timestamp:       time.Now().UTC(),
		Monitor:         monitor,
	}

	go s.processor.ProcessResponse(context.WithoutCancel(r.Context()), response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HttpCommonSuccess{Message: "success"})
}
