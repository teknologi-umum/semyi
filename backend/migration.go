package main

import (
	"bufio"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Migrate(db *sql.DB, ctx context.Context, directionUp bool) error {
	dir, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var files []string

	for _, file := range dir {
		if file.IsDir() {
			continue
		}

		files = append(files, file.Name())
	}

	sort.SliceStable(files, func(i, j int) bool {
		// Gather the first 14 characters of the file name. Then we can parse it as YYYYMMDDHHmmss and sort by that.
		firstFileName := files[i][:14]
		secondFileName := files[j][:14]

		firstTime, _ := strconv.ParseInt(firstFileName, 10, 64)
		secondTime, _ := strconv.ParseInt(secondFileName, 10, 64)

		return firstTime < secondTime
	})

	var migrationScripts []string
	for _, file := range files {
		content, err := migrationFiles.Open("migrations/" + file)
		if err != nil {
			if content != nil {
				_ = content.Close()
			}

			return fmt.Errorf("failed to read migration file: %w", err)
		}

		var contentAccumulator strings.Builder
		var foundStartMarker = false
		scanner := bufio.NewScanner(content)
		scanner.Split(bufio.ScanLines)
		if directionUp {
			// Read from the line that has "-- +goose Up" until the first occurrence of "-- +goose StatementEnd"
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "-- +goose Up" {
					foundStartMarker = true
					continue
				}

				if foundStartMarker {
					if line == "-- +goose StatementEnd" {
						break
					}

					contentAccumulator.WriteString(line)
					contentAccumulator.WriteString("\n")
				}
			}
		} else {
			// Read from the line that has "-- +goose Down" until the first occurrence of "-- +goose StatementEnd"
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "-- +goose Down" {
					foundStartMarker = true
					continue
				}

				if foundStartMarker {
					if line == "-- +goose StatementEnd" {
						break
					}

					contentAccumulator.WriteString(line)
					contentAccumulator.WriteString("\n")
				}
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed to read migration file: %w", err)
		}

		err = content.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close file")
		}

		migrationScripts = append(migrationScripts, contentAccumulator.String())
	}

	c, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer func() {
		err := c.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close connection")
		}
	}()

	tx, err := c.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for _, migrationScript := range migrationScripts {
		// We split everything by `;` and execute each statement in the transaction.
		for statement := range strings.SplitSeq(migrationScript, ";") {
			if strings.TrimSpace(statement) == "" {
				continue
			}

			_, err = tx.ExecContext(
				ctx,
				statement,
			)
			if err != nil {
				if e := tx.Rollback(); e != nil {
					return fmt.Errorf("failed to rollback transaction: %w (%s)", e, err.Error())
				}

				return fmt.Errorf("failed to execute migration script: %w", err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
