package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/allegro/bigcache/v3"
	_ "modernc.org/sqlite"
)

type Deps struct {
	DB              *sql.DB
	Queue           *Queue
	Cache           *bigcache.BigCache
	DefaultTimeout  int
	DefaultInterval int
}

func main() {
	// Read environment variables
	configPath, ok := os.LookupEnv("CONFIG_PATH")
	if !ok {
		configPath = "../config.json"
	}

	dbPath, ok := os.LookupEnv("DB_PATH")
	if !ok {
		dbPath = "../db.sqlite3"
	}

	staticPath, ok := os.LookupEnv("STATIC_PATH")
	if !ok {
		staticPath = "../frontend/dist"
	}

	defaultInterval, ok := os.LookupEnv("DEFAULT_INTERVAL")
	if !ok {
		defaultInterval = "30"
	}

	defaultTimeout, ok := os.LookupEnv("DEFAULT_TIMEOUT")
	if !ok {
		defaultTimeout = "10"
	}

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "5000"
	}

	defTimeout, err := strconv.Atoi(defaultTimeout)
	if err != nil {
		log.Fatalf("Failed to parse default timeout: %v", err)
	}

	defInterval, err := strconv.Atoi(defaultInterval)
	if err != nil {
		log.Fatalf("Failed to parse default interval: %v", err)
	}

	// Read configuration file
	config, err := ReadConfigurationFile(configPath)
	if err != nil {
		log.Fatalf("failed to read configuration file: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(time.Hour * 24))
	if err != nil {
		log.Fatalf("failed to create cache: %v", err)
	}
	defer cache.Close()

	deps := &Deps{
		DB:              db,
		Cache:           cache,
		Queue:           NewQueue(),
		DefaultTimeout:  defTimeout,
		DefaultInterval: defInterval,
	}

	// Create a new worker
	for _, endpoint := range config {
		worker, err := deps.NewWorker(&endpoint)
		if err != nil {
			log.Fatalf("Failed to create worker: %v", err)
		}

		// register endpoint url into cache
		err = deps.Cache.Append("endpoint:urls", []byte(endpoint.URL+","))
		if err != nil {
			log.Fatalf("Failed to register endpoint url into cache: %v", err)
		}

		go func() {
			worker.Run()
		}()
	}

	// Dump snapshot every 5 seconds
	go func() {
		for {
			deps.Queue.Lock()

			if len(deps.Queue.Items) > 0 {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

				err := deps.WriteSnapshot(ctx, deps.Queue.Items)
				if err != nil {
					cancel()
					log.Printf("Failed to write snapshot: %v", err)
				}

				if err == nil {
					deps.Queue.Items = []Response{}
				}

				cancel()
			}

			deps.Queue.Unlock()
			time.Sleep(time.Second * 5)
		}
	}()

	server := deps.NewServer(port, staticPath)
	go func() {
		// Start the server
		if e := server.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", e)
		}
	}()

	// Listen for SIGKILL and SIGTERM
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}

	deps.Queue.Lock()

	err = deps.WriteSnapshot(ctx, deps.Queue.Items)
	if err != nil {
		log.Printf("Failed to write snapshot: %v", err)
	}

	deps.Queue.Unlock()
}
