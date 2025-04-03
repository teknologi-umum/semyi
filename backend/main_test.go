package main_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	main "semyi"

	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/rs/zerolog/log"
)

var database *sql.DB

func TestMain(m *testing.M) {
	// Setup
	var err error
	database, err = sql.Open("duckdb", "")
	if err != nil {
		log.Error().Err(err).Msg("failed to open database connection")
		os.Exit(1)
		return
	}

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
