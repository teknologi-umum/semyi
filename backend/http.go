package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"github.com/unrolled/secure"
)

type Server struct {
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
	CentralBroker           *Broker[MonitorHistorical]
	IncidentWriter          *IncidentWriter
	MonitorList             []Monitor

	ApiKey string
}

func NewServer(config ServerConfig) *http.Server {
	server := &Server{
		historicalReader: config.MonitorHistoricalReader,
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
		AllowedMethods: []string{"GET", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	})

	api := chi.NewRouter()
	api.Use(corsMiddleware.Handler)
	api.Get("/api/overview", server.snapshotOverview)
	api.Get("/api/by", server.snapshotBy)
	api.Get("/api/static", server.staticSnapshot)
	api.Post("/api/incident", server.submitIncindent)

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
		w.WriteHeader(http.StatusPreconditionFailed)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "not flusher"}`))
		return
	}

	subscriber, err := NewSubscriber(s.centralBroker, monitorIds...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		errBytes, err := json.Marshal(map[string]string{"error": fmt.Errorf("failed to subscribe to endpoints: %s", err).Error()})
		if err != nil {
			w.Write([]byte(`{"error": "internal server error"}`))
			return
		}
		w.Write(errBytes)
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
		w.WriteHeader(http.StatusPreconditionFailed)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "not flusher"}`))
		return
	}

	ids := r.URL.Query().Get("ids")
	if ids == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error":"ids is required"}`))
		return
	}

	wantedMonitorIds := strings.Split(ids, ",")

	for _, id := range wantedMonitorIds {
		if !slices.Contains(monitorIds, id) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "id is not in the list of monitors"}`))
			return
		}
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	sub, err := NewSubscriber(s.centralBroker, wantedMonitorIds...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		errBytes, err := json.Marshal(map[string]string{"error": fmt.Errorf("failed to subscribe to endpoints: %s", err).Error()})
		if err != nil {
			w.Write([]byte(`{"error": "internal server error"}`))
			return
		}
		w.Write(errBytes)
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
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "id is required"}`))
		return
	}

	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "hourly"
	}

	if interval != "hourly" && interval != "daily" && interval != "raw" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "interval must be hourly, daily, or raw"}`))
		return
	}

	if !slices.Contains(monitorIds, monitorId) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "id is not in the list of monitors"}`))
		return
	}

	var err error
	var monitor Monitor
	var monitorHistorical []MonitorHistorical
	switch interval {
	case "raw":
		monitorHistorical, err = s.historicalReader.ReadRawHistorical(r.Context(), monitorId)
		if err != nil {
			// TODO: Handle error properly
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		break
	case "hourly":
		monitorHistorical, err = s.historicalReader.ReadHourlyHistorical(r.Context(), monitorId)
		if err != nil {
			// TODO: Handle error properly
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		break
	case "daily":
		monitorHistorical, err = s.historicalReader.ReadDailyHistorical(r.Context(), monitorId)
		if err != nil {
			// TODO: Handle error properly
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		break
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "interval must be hourly, daily, or raw"}`))
		return
	}

	// Acquire monitor metadata
	for _, m := range s.monitors {
		if m.UniqueID == monitorId {
			monitor = m
			break
		}
	}

	data, err := json.Marshal(map[string]any{
		"metadata":   monitor,
		"historical": monitorHistorical,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		errBytes, err := json.Marshal(map[string]string{"error": err.Error()})
		if err != nil {
			w.Write([]byte(`{"error": "internal server error"}`))
			return
		}
		w.Write(errBytes)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *Server) submitIncindent(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("x-api-key")
	if apiKey == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "api key is required"}`))
		return
	} else {
		if apiKey != s.apiKey {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": "api key is invalid"}`))
			return
		}
	}

	decoder := json.NewDecoder(r.Body)
	var body Incident
	if err := decoder.Decode(&body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		errBytes, marshalErr := json.Marshal(map[string]string{
			"error": err.Error(),
		})
		if marshalErr != nil {
			log.Error().Stack().Err(err).Msg("failed to marshal json")
			w.Write([]byte(`{"error": "internal server error"}`))
			return
		}
		w.Write(errBytes)
		return
	}
	defer r.Body.Close()

	if err := body.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		errBytes, marshalErr := json.Marshal(map[string]string{
			"error": err.Error(),
		})
		if marshalErr != nil {
			log.Error().Stack().Err(err).Msg("failed to marshal json")
			w.Write([]byte(`{"error": "internal server error"}`))
			return
		}
		w.Write(errBytes)
		return
	}

	err := s.incidentWriter.Write(r.Context(), body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		errBytes, err := json.Marshal(map[string]string{
			"error": err.Error(),
		})
		if err != nil {
			log.Error().Stack().Err(err).Msg("failed to marshal json")
			w.Write([]byte(`{"error": "internal server error"}`))
			return
		}
		w.Write(errBytes)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "success"}`))
}
