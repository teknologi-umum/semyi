package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type WebhookRequest struct {
	Endpoint        string `json:"endpoint"`
	Status          string `json:"status"`
	StatusCode      int    `json:"statusCode"`
	RequestDuration int64  `json:"requestDuration"`
	Timestamp       int64  `json:"timestamp"`
}

func main() {
	router := http.NewServeMux()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body WebhookRequest
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		w.WriteHeader(http.StatusAccepted)
		log.Printf("Received webhook: %+v", body)
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Printf("server listening on http://localhost%s", server.Addr)
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("failed to start server: %s", err)
	}
}
