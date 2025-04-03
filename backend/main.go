package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/marcboeker/go-duckdb"
	"github.com/rs/zerolog/log"
)

var (
	DefaultInterval int = 30
	DefaultTimeout  int = 10
	monitorIds      []string
)

func main() {
	// Read environment variables
	configPath, ok := os.LookupEnv("CONFIG_PATH")
	if !ok {
		configPath = "../config.json"
	}

	// possible options: development, production
	environment, ok := os.LookupEnv("ENVIRONMENT")
	if !ok {
		environment = "development"
	} else {
		if environment != "development" && environment != "production" {
			log.Fatal().Msg("Invalid environment. Possible options: development, production")
		}
	}

	hostname, ok := os.LookupEnv("HOSTNAME")
	if !ok {
		hostname = "0.0.0.0"
	}

	dbPath, ok := os.LookupEnv("DB_PATH")
	if !ok {
		dbPath = "../db.duckdb"
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

	apiKey, ok := os.LookupEnv("API_KEY")
	if !ok {
		log.Warn().Msg("API_KEY is not set")
	}

	telegramChatID, ok := os.LookupEnv("TELEGRAM_CHAT_ID")
	if !ok {
		log.Warn().Msg("TELEGRAM_CHAT_ID is not set")
	}

	telegramUrl, ok := os.LookupEnv("TELEGRAM_URL is not set")
	if !ok {
		log.Warn().Msg("TELEGRAM_URL is not set")
	}

	if os.Getenv("ENV") == "" {
		err := os.Setenv("ENV", "development")
		if err != nil {
			log.Fatal().Err(err).Msg("Error setting ENV")
		}
	}

	// Read configuration file
	config, err := ReadConfigurationFile(configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read configuration file")
	}

	DefaultTimeout, err = strconv.Atoi(defaultTimeout)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse default timeout")
	}

	DefaultInterval, err = strconv.Atoi(defaultInterval)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse default interval")
	}

	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to close database")
		}
	}(db)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = Migrate(db, ctx, true)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to migrate database")
	}

	processor := &Processor{
		telegramAlertProvider: NewTelegramAlertProvider(TelegramProviderConfig{
			Url:    telegramUrl,
			ChatID: telegramChatID,
		}),
	}

	// Create a new worker
	for _, monitor := range config.Monitors {
		monitorIds = append(monitorIds, monitor.UniqueID)

		worker, err := NewWorker(monitor, processor)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create worker")
		}

		log.Info().Str("UniqueID", monitor.UniqueID).Str("Name", monitor.Name).Msg("Registered monitor")

		go func(worker *Worker) {
			defer func() {
				if r := recover(); r != nil {
					log.Warn().Msgf("[Running worker] Recovered from panic: %v", r)
				}
			}()

			worker.Run()
		}(worker)
	}

	monitorHistoricalReader := NewMonitorHistoricalReader(db)
	monitorHistoricalWriter := NewMonitorHistoricalWriter(db)

	aggregateWorker := NewAggregateWorker(monitorIds, monitorHistoricalReader, monitorHistoricalWriter)

	go aggregateWorker.RunDailyAggregate()
	go aggregateWorker.RunHourlyAggregate()

	server := NewServer(ServerConfig{
		SSLRedirect:             false,
		Environment:             environment,
		Hostname:                hostname,
		Port:                    port,
		StaticPath:              staticPath,
		MonitorHistoricalReader: monitorHistoricalReader,
		MonitorHistoricalWriter: monitorHistoricalWriter,
		CentralBroker:           &Broker[MonitorHistorical]{},
		IncidentWriter:          NewIncidentWriter(db),
		MonitorList:             config.Monitors,

		ApiKey: apiKey,
	})
	go func() {
		// Listen for SIGKILL and SIGTERM
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
		<-signalChan

		log.Info().Msg("Shutting down server...")
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		err = server.Shutdown(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to shutdown server")
		}
	}()

	// Start the server
	log.Printf("Starting server on port %s", port)
	if e := server.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
