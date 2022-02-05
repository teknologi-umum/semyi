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
	})

	api := chi.NewRouter()
	api.Use(corsMiddleware.Handler)
	api.Get("/overview", func(rw http.ResponseWriter, r *http.Request) {
		err := d.snapshotOverview(rw, r)
		if err != nil {
			log.Printf("failed to get snapshot: %s", err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Header().Set("Content-Type", "application/json")
			rw.Write([]byte(`{"error": "` + err.Error() + `"}`))
		}
	})
	api.Get("/by", func(rw http.ResponseWriter, r *http.Request) {
		err := d.snapshotBy(rw, r)
		if err != nil {
			log.Printf("failed to get snapshot by url: %s", err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Header().Set("Content-Type", "application/json")
			rw.Write([]byte(`{"error": "` + err.Error() + `"}`))
		}
	})
	api.Get("/static", func(rw http.ResponseWriter, r *http.Request) {
		err := d.staticSnapshot(rw, r)
		if err != nil {
			log.Printf("failed to serve static: %s", err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Header().Set("Content-Type", "application/json")
			rw.Write([]byte(`{"error": "` + err.Error() + `"}`))
		}
	})

	r := chi.NewRouter()
	r.Use(secureMiddleware.Handler)
	r.Mount("/api", api)
	r.Handle("/", http.FileServer(http.Dir(staticPath)))

	return &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
}

func (d *Deps) snapshotOverview(w http.ResponseWriter, r *http.Request) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("not flusher")
	}

	endpointsBytes, err := d.Cache.Get("endpoint:urls")
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	go func() {
		endpoints := strings.Split(string(endpointsBytes), ",")
		sub, err := d.NewSubscriber(endpoints...)
		if err != nil {
			log.Printf("failed to subscribe to endpoints: %s", err)
			return
		}
	
		for data := range sub.Listen(r.Context()) {
			marshaled, err := json.Marshal(data)
			if err != nil {
				log.Printf("failed to marshal data: %s", err)
				continue
			}

			_, err = w.Write([]byte("data: " + string(marshaled) + "\n\n"))
			if err != nil {
				log.Printf("failed to write data: %s", err)
				continue
			}

			flusher.Flush()
		}
	}()
	
	return nil
}

func (d *Deps) snapshotBy(w http.ResponseWriter, r *http.Request) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("not flusher")
	}

	url := r.URL.Query().Get("url")
	if url == "" {
		return fmt.Errorf("url is none")
	}

	endpointsBytes, err := d.Cache.Get("endpoint:urls")
	if err != nil {
		return err
	}

	if !strings.Contains(string(endpointsBytes), url) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "url is not in the list of endpoints"}`))
		return nil
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	go func() {
		endpoints := strings.Split(url, ",")
		sub, err := d.NewSubscriber(endpoints...)
		if err != nil {
			log.Printf("failed to subscribe to endpoints: %s", err)
			return
		}

		for data := range sub.Listen(r.Context()) {
			marshaled, err := json.Marshal(data)
			if err != nil {
				log.Printf("failed to marshal data: %s", err)
				continue
			}

			_, err = w.Write([]byte("data: " + string(marshaled) + "\n\n"))
			if err != nil {
				log.Printf("failed to write data: %s", err)
				continue
			}

			flusher.Flush()
		}
	}()

	return nil
}

func (d *Deps) staticSnapshot(w http.ResponseWriter, r *http.Request) error {
	url := r.URL.Query().Get("url")
	if url == "" {
		return fmt.Errorf("url is none")
	}

	endpointsBytes, err := d.Cache.Get("endpoint:urls")
	if err != nil {
		return err
	}
	
	endpoints := strings.Split(string(endpointsBytes), ",")

	if !contains(url, endpoints) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "url is not in the list of endpoints"}`))
		return nil
	}

	// acquire endpoint metadata from cache
	endpointBytes, err := d.Cache.Get("endpoint:"+url)
	if err != nil {
		return err
	}

	var endpoint Endpoint
	err = json.Unmarshal(endpointBytes, &endpoint)
	if err != nil {
		return err
	}

	c, err := d.DB.Conn(r.Context())
	if err != nil {
		return err
	}
	defer c.Close()

	tx, err := c.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelReadUncommitted, ReadOnly: true})
	if err != nil {
		return err
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
		return err
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
			return err
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
		return err
	}

	data, err := json.Marshal(snapshots)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	return err
}
