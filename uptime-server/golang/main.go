package main

import (
	"log"
	"math/rand/v2"
	"net/http"
	"time"
)

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok := rand.Float64() < 0.5
		latency := rand.Float64() * 100
		time.Sleep(time.Duration(latency) * time.Millisecond)
		if ok {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("ERROR"))
		}
	})

	server := &http.Server{
		Addr:    ":9000",
		Handler: handler,
	}

	log.Printf("server listening on http://localhost%s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
