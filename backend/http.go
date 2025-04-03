package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
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

		apiKey: config.ApiKey,
	}

	secureMiddleware := secure.New(secure.Options{
		BrowserXssFilter:   true,
		ContentTypeNosniff: true,
		SSLRedirect:        config.SSLRedirect,
		IsDevelopment:      config.Environment == "development",
	})

	corsMiddleware := cors.New(cors.Options{
		Debug:          config.Environment == "development",
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	})

	api := chi.NewRouter()
	api.Use(corsMiddleware.Handler)
	api.Get("/api/overview", server.snapshotOverview)
	api.Get("/api/by", server.snapshotBy)
	api.Get("/api/static", server.staticSnapshot)
	api.Post("/api/incident", server.submitIncindent)
	api.Get("/api/push/{monitor_id}", server.pushHealthcheck)

	r := chi.NewRouter()
	r.Use(secureMiddleware.Handler)
	r.Handle("/api/*", corsMiddleware.Handler(api))
	r.Handle("/", http.FileServer(http.Dir(config.StaticPath)))

	return &http.Server{
		Addr:    net.JoinHostPort(config.Hostname, config.Port),
		Handler: r,
	}
}

func (s *Server) snapshotOverview(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPreconditionFailed)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "not flusher"})
		return
	}

	subscriber, err := NewSubscriber(s.centralBroker, monitorIds...)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: fmt.Sprintf("failed to subscribe to endpoints: %s", err)})
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case data := <-subscriber.Listen(r.Context()):
			marshaled, err := json.Marshal(data)
			if err != nil {
				log.Printf("failed to marshal data: %s", err)
			}

			_, err = w.Write([]byte("data: " + string(marshaled) + "\n\n"))
			if err != nil {
				log.Printf("failed to write data: %s", err)
			}

			flusher.Flush()
		default:
			time.Sleep(time.Millisecond * 10)
		}
	}
}

func (s *Server) snapshotBy(w http.ResponseWriter, r *http.Request) {
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

	sub, err := NewSubscriber(s.centralBroker, wantedMonitorIds...)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: fmt.Sprintf("failed to subscribe to endpoints: %s", err)})
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case data := <-sub.Listen(r.Context()):
			marshaled, err := json.Marshal(data)
			if err != nil {
				log.Printf("failed to marshal data: %s", err)
			}

			_, err = w.Write([]byte("data: " + string(marshaled) + "\n\n"))
			if err != nil {
				log.Printf("failed to write data: %s", err)
			}

			flusher.Flush()
		default:
			time.Sleep(time.Millisecond * 10)
		}
	}
}

func (s *Server) staticSnapshot(w http.ResponseWriter, r *http.Request) {
	monitorId := r.URL.Query().Get("id")
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
		monitorHistorical, err = s.historicalReader.ReadRawHistorical(r.Context(), monitorId)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(HttpCommonError{Error: fmt.Sprintf("failed to read raw historical data: %s", err)})
			return
		}
	case "hourly":
		monitorHistorical, err = s.historicalReader.ReadHourlyHistorical(r.Context(), monitorId)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(HttpCommonError{Error: fmt.Sprintf("failed to read hourly historical data: %s", err)})
			return
		}
	case "daily":
		monitorHistorical, err = s.historicalReader.ReadDailyHistorical(r.Context(), monitorId)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(HttpCommonError{Error: fmt.Sprintf("failed to read daily historical data: %s", err)})
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
	apiKey := r.Header.Get("x-api-key")
	if apiKey == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "api key is required"})
		return
	} else {
		if apiKey != s.apiKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "api key is invalid"})
			return
		}
	}

	decoder := json.NewDecoder(r.Body)
	var body Incident
	if err := decoder.Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: err.Error()})
		return
	}
	defer r.Body.Close()

	if err := body.Validate(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: err.Error()})
		return
	}

	err := s.incidentWriter.Write(r.Context(), body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HttpCommonSuccess{Message: "success"})
}

func (s *Server) pushHealthcheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "Method not allowed"})
		return
	}

	// Get monitor ID from URL path
	monitorID := chi.URLParam(r, "monitor_id")
	if monitorID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "monitor_id is required"})
		return
	}

	// Trim monitorID
	monitorID = strings.TrimSpace(monitorID)

	// Validate monitor ID exists
	if !slices.Contains(monitorIds, monitorID) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "Invalid monitor_id"})
		return
	}

	// Get query parameters
	status := r.URL.Query().Get("status")
	pingStr := r.URL.Query().Get("ping")

	// Convert ping to latency (in milliseconds)
	var latency int64
	if pingStr != "" {
		ping, err := strconv.ParseFloat(pingStr, 64)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "Invalid ping value"})
			return
		}
		latency = int64(ping * 1000) // Convert seconds to milliseconds
	}

	// Convert status string to MonitorStatus
	var monitorStatus MonitorStatus
	switch strings.ToLower(status) {
	case "up":
		monitorStatus = MonitorStatusSuccess
	case "down":
		monitorStatus = MonitorStatusFailure
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "Invalid status value"})
		return
	}

	// Create a new monitor historical record
	historical := MonitorHistorical{
		MonitorID: monitorID,
		Status:    monitorStatus,
		Latency:   latency,
		Timestamp: time.Now().UTC(),
	}

	// Validate the historical record
	if valid, err := historical.Validate(); !valid || err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "Invalid healthcheck data"})
		return
	}

	// Write to storage
	if err := s.historicalWriter.Write(r.Context(), historical); err != nil {
		log.Error().Err(err).Msg("Failed to write healthcheck result")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(HttpCommonError{Error: "Failed to write healthcheck result"})
		return
	}

	// Publish to broker for real-time updates
	s.centralBroker.Publish(monitorID, &BrokerMessage[MonitorHistorical]{
		Header: make(map[string]string),
		Body:   historical,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HttpCommonSuccess{Message: "OK"})
}
