package main_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"os"
	"strings"
	"testing"
	"time"

	main "semyi"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/marcboeker/go-duckdb/v2"
	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/rs/zerolog/log"
)

var database *sql.DB

func TestMain(m *testing.M) {
	// Setup
	var err error
	databasePath := os.Getenv("DATABASE_URL")
	var connector driver.Connector
	if strings.HasPrefix(databasePath, "clickhouse://") || strings.HasPrefix(databasePath, "http://") {
		clickHouseOptions, err := clickhouse.ParseDSN(databasePath)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse clickhouse DSN")
			os.Exit(1)
			return
		}

		connector = clickhouse.Connector(clickHouseOptions)
	} else {
		connector, err = duckdb.NewConnector(databasePath, func(execer driver.ExecerContext) error {
			return nil
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to create duckdb connector")
			os.Exit(1)
			return
		}
	}

	database = sql.OpenDB(connector)
	database.SetConnMaxIdleTime(time.Minute)
	database.SetConnMaxLifetime(time.Minute * 10)

	// Migrate database
	err = main.Migrate(database, context.Background(), true)
	if err != nil {
		log.Error().Err(err).Msg("failed to migrate database")
		os.Exit(1)
		return
	}

	// Run tests
	exitCode := m.Run()

	// Teardown
	err = database.Close()
	if err != nil {
		log.Warn().Err(err).Msg("failed to close database connection")
	}

	os.Exit(exitCode)
}
