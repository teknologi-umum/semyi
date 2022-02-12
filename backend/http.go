package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
)

func (d *Deps) NewServer(port, staticPath string) *http.Server {
	secureMiddleware := secure.New(secure.Options{
		BrowserXssFilter:   true,
		ContentTypeNosniff: true,
		SSLRedirect:        os.Getenv("ENV") == "production",
		IsDevelopment:      os.Getenv("ENV") == "development",
	})

	corsMiddleware := cors.New(cors.Options{
		Debug:          os.Getenv("ENV") == "development",
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	})

	api := chi.NewRouter()
	api.Use(corsMiddleware.Handler)
	api.Get("/overview", d.snapshotOverview)
	api.Get("/by", d.snapshotBy)
	api.Get("/static", d.staticSnapshot)

	r := chi.NewRouter()
	r.Use(secureMiddleware.Handler)
	r.Mount("/api", api)
	r.Handle("/", http.FileServer(http.Dir(staticPath)))

	return &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
}

func (d *Deps) snapshotOverview(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusPreconditionFailed)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "not flusher"}`))
		return
	}

	endpointsBytes, err := d.Cache.Get("endpoint:urls")
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

	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)

	endpoints := strings.Split(string(endpointsBytes), ",")
	sub, err := d.NewSubscriber(endpoints...)
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
		}
	}

}

func (d *Deps) snapshotBy(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusPreconditionFailed)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "not flusher"}`))
		return
	}

	url := r.URL.Query().Get("url")
	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "url is required"}`))
		return
	}

	endpointsBytes, err := d.Cache.Get("endpoint:urls")
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

	if !strings.Contains(string(endpointsBytes), url) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "url is not in the list of endpoints"}`))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	endpoints := strings.Split(url, ",")
	sub, err := d.NewSubscriber(endpoints...)
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
		}
	}

}

func (d *Deps) staticSnapshot(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "url is required"}`))
		return
	}

	endpointsBytes, err := d.Cache.Get("endpoint:urls")
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

	endpoints := strings.Split(string(endpointsBytes), ",")

	if !contains(url, endpoints) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "url is not in the list of endpoints"}`))
		return
	}

	// acquire endpoint metadata from cache
	endpointBytes, err := d.Cache.Get("endpoint:" + url)
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

	var endpoint Endpoint
	err = json.Unmarshal(endpointBytes, &endpoint)
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

	c, err := d.DB.Conn(r.Context())
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
	defer c.Close()

	tx, err := c.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelReadUncommitted, ReadOnly: true})
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

	rows, err := tx.QueryContext(
		r.Context(),
		`SELECT
			url,
			timeout,
			interval,
			status_code,
			request_duration,
			created_at
		FROM 
			snapshot 
		WHERE 
			url = ?
		ORDER BY 
			created_at DESC 
		LIMIT 100`,
		url,
	)
	if err != nil {
		tx.Rollback()
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
	defer rows.Close()

	var snapshots []Response
	for rows.Next() {
		var snapshot Response
		err := rows.Scan(
			&snapshot.URL,
			&snapshot.Timeout,
			&snapshot.Interval,
			&snapshot.StatusCode,
			&snapshot.RequestDuration,
			&snapshot.Timestamp,
		)
		if err != nil {
			tx.Rollback()
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

		snapshot.Name = endpoint.Name
		snapshot.Description = endpoint.Description
		snapshot.Method = endpoint.Method
		snapshot.Headers = endpoint.Headers

		snapshots = append(snapshots, snapshot)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
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

	data, err := json.Marshal(snapshots)
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
