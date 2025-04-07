package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/aldy505/sentry-integration/httpclient"
	"github.com/aldy505/sentry-integration/sqltracer"
	"github.com/getsentry/sentry-go"
	"github.com/marcboeker/go-duckdb/v2"
	"github.com/rs/zerolog/log"
)

var (
	DefaultInterval int = 30
	DefaultTimeout  int = 10
	monitorIds      []string
	release         string
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

	enableDumpFailureResponse := false
	if value, ok := os.LookupEnv("ENABLE_DUMP_FAILURE_RESPONSE"); ok {
		enableDumpFailureResponse, _ = strconv.ParseBool(value)
	}

	sentrySampleRate := 1.0
	if os.Getenv("BACKEND_SENTRY_SAMPLE_RATE") != "" {
		var err error = nil
		sentrySampleRate, err = strconv.ParseFloat(os.Getenv("BACKEND_SENTRY_SAMPLE_RATE"), 64)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse sentry sample rate")
		}
	}

	sentryTracesSampleRate := 1.0
	if os.Getenv("BACKEND_SENTRY_TRACES_SAMPLE_RATE") != "" {
		var err error = nil
		sentryTracesSampleRate, err = strconv.ParseFloat(os.Getenv("BACKEND_SENTRY_TRACES_SAMPLE_RATE"), 64)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse sentry traces sample rate")
		}
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("BACKEND_SENTRY_DSN"),
		SampleRate:       sentrySampleRate,
		EnableTracing:    true,
		TracesSampleRate: sentryTracesSampleRate,
		Release:          release,
		Environment:      environment,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize sentry")
	}
	defer sentry.Flush(time.Second * 10)

	httpClient := &http.Client{
		Transport: httpclient.NewSentryRoundTripper(nil, nil),
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

	var connector driver.Connector
	// If the dbPath has `clickhouse://` or `http://` prefix, we use clickhouse by parsing the DSN and using the clickhouse-go driver
	// to create a new `database/sql` compatible connector. Otherwise, we use the duckdb driver by using the `dbPath` as is.
	if strings.HasPrefix(dbPath, "clickhouse://") || strings.HasPrefix(dbPath, "http://") {
		clickHouseOptions, err := clickhouse.ParseDSN(dbPath)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse clickhouse DSN")
		}

		connector = sqltracer.NewSentrySQLConnector(
			clickhouse.Connector(clickHouseOptions),
			sqltracer.WithDatabaseSystem("clickhouse"),
		)
	} else {
		connector, err = duckdb.NewConnector(dbPath, func(execer driver.ExecerContext) error {
			return nil
		})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create duckdb connector")
		}

		connector = sqltracer.NewSentrySQLConnector(
			connector,
			sqltracer.WithDatabaseSystem("duckdb"),
			sqltracer.WithDatabaseName(dbPath),
		)
	}
	var db *sql.DB = sql.OpenDB(connector)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(time.Minute * 30)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Warn().Err(err).Msg("failed to close database")
		}
	}(db)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = Migrate(db, ctx, true)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to migrate database")
	}

	monitorHistoricalReader := NewMonitorHistoricalReader(db)
	monitorHistoricalWriter := NewMonitorHistoricalWriter(db)
	centralBroker := NewBroker[MonitorHistorical]()

	aggregateWorker := NewAggregateWorker(monitorIds, monitorHistoricalReader, monitorHistoricalWriter)

	processor := &Processor{
		HistoricalWriter: monitorHistoricalWriter,
		HistoricalReader: monitorHistoricalReader,
		CentralBroker:    centralBroker,
	}

	// Initialize alert providers if enabled
	if config.Alerting.Telegram.Enabled && config.Alerting.Telegram.URL != "" && config.Alerting.Telegram.ChatID != "" {
		processor.TelegramAlertProvider = NewTelegramAlertProvider(TelegramProviderConfig{
			Url:        config.Alerting.Telegram.URL,
			ChatID:     config.Alerting.Telegram.ChatID,
			HttpClient: httpClient,
		})
	}

	if config.Alerting.Discord.Enabled && config.Alerting.Discord.WebhookURL != "" {
		processor.DiscordAlertProvider = NewDiscordAlertProvider(DiscordProviderConfig{
			WebhookURL: config.Alerting.Discord.WebhookURL,
			HttpClient: httpClient,
		})
	}

	if config.Alerting.HTTP.Enabled && config.Alerting.HTTP.WebhookURL != "" {
		processor.HTTPAlertProvider = NewHTTPAlertProvider(HTTPProviderConfig{
			WebhookURL: config.Alerting.HTTP.WebhookURL,
			HttpClient: httpClient,
		})
	}

	if config.Alerting.Slack.Enabled && config.Alerting.Slack.WebhookURL != "" {
		processor.SlackAlertProvider = NewSlackAlertProvider(SlackProviderConfig{
			WebhookURL: config.Alerting.Slack.WebhookURL,
			HttpClient: httpClient,
		})
	}

	// Create a new worker
	for _, monitor := range config.Monitors {
		monitorIds = append(monitorIds, monitor.UniqueID)

		worker, err := NewWorker(monitor, processor, enableDumpFailureResponse)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create worker")
		}

		log.Info().Str("UniqueID", monitor.UniqueID).Str("Name", monitor.Name).Msg("Registered monitor")

		go func(worker *Worker) {
			defer func() {
				if r, ok := recover().(error); ok && r != nil {
					sentry.CaptureException(r)
					log.Warn().Str("monitor_id", worker.monitor.UniqueID).Msgf("[Running worker] Recovered from panic: %v", r)
				}
			}()

			worker.Run()
		}(worker)
	}

	go aggregateWorker.RunDailyAggregate()
	go aggregateWorker.RunHourlyAggregate()

	// Initialize cleanup worker
	cleanupWorker := NewCleanupWorker(db, config.RetentionPeriod)
	go cleanupWorker.Run(context.Background())

	server := NewServer(ServerConfig{
		SSLRedirect:             false,
		Environment:             environment,
		Hostname:                hostname,
		Port:                    port,
		StaticPath:              staticPath,
		MonitorHistoricalReader: monitorHistoricalReader,
		MonitorHistoricalWriter: monitorHistoricalWriter,
		CentralBroker:           centralBroker,
		IncidentWriter:          NewIncidentWriter(db),
		MonitorList:             config.Monitors,
		ApiKey:                  apiKey,
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
			sentry.CaptureException(err)
			log.Fatal().Err(err).Msg("Failed to shutdown server")
		}
	}()

	// Start the server
	log.Info().Msgf("Starting server on port %s", port)
	if e := server.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
