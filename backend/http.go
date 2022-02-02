package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

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
		AllowedOrigins: []string{},
		AllowedMethods: []string{"GET", "OPTIONS"},
	})

	api := http.NewServeMux()
	api.HandleFunc("/overview", func(rw http.ResponseWriter, r *http.Request) {
		err := d.snapshotOverview(rw, r)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Header().Set("Content-Type", "application/json")
			rw.Write([]byte(`{"error": "` + err.Error() + `"}`))
		}
	})
	api.HandleFunc("/by", func(rw http.ResponseWriter, r *http.Request) {
		err := d.fetchSnapshot(rw, r)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Header().Set("Content-Type", "application/json")
			rw.Write([]byte(`{"error": "` + err.Error() + `"}`))
		}
	})

	r := http.NewServeMux()
	r.Handle("/", http.FileServer(http.Dir(staticPath)))
	r.Handle("/api", corsMiddleware.Handler(secureMiddleware.Handler(api)))

	return &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
}

func (d *Deps) snapshotOverview(w http.ResponseWriter, r *http.Request) error {
	endpointsBytes, err := d.Cache.Get("endpoint:urls")
	if err != nil {
		return err
	}

	endpoints := strings.Split(string(endpointsBytes), ",")

	var snapshots []Response

	for _, endpoint := range endpoints {
		if endpoint == "" {
			continue
		}

		snapshotBytes, err := d.Cache.Get(endpoint)
		if err != nil {
			return err
		}

		var snapshot Response
		err = json.Unmarshal(snapshotBytes, &snapshot)
		if err != nil {
			return err
		}

		snapshots = append(snapshots, snapshot)
	}

	data, err := json.Marshal(snapshots)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Write([]byte("data: " + string(data) + "\n\n"))
	return nil
}

func (d *Deps) fetchSnapshot(w http.ResponseWriter, r *http.Request) error {
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

	c, err := d.DB.Conn(r.Context())
	if err != nil {
		return err
	}
	defer c.Close()

	tx, err := c.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true})
	if err != nil {
		return err
	}

	rows, err := tx.QueryContext(
		r.Context(),
		`SELECT
			name,
			url,
			description,
			timeout,
			interval,
			status_code,
			request_duration,
			created_at
		FROM 
			snapshots 
		WHERE 
			url = $1 
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
			&snapshot.Name,
			&snapshot.URL,
			&snapshot.Description,
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

		snapshots = append(snapshots, snapshot)
	}

	data, err := json.Marshal(snapshots)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Write([]byte("data: " + string(data) + "\n\n"))
	return nil
}
