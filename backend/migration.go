package main

import (
	"context"
	"database/sql"
)

func Migrate(db *sql.DB, ctx context.Context) error {
	c, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	tx, err := c.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`CREATE TABLE IF NOT EXISTS snapshot (
			url varchar(255) NOT NULL,
			timeout integer DEFAULT 10,
			interval integer DEFAULT 30,
			status_code integer NOT NULL,
			request_duration integer NOT NULL,
			created_at timestamp NOT NULL
		)`,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`CREATE INDEX IF NOT EXISTS snapshot_url_idx ON snapshot (url)`,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`CREATE INDEX IF NOT EXISTS snapshot_created_at_idx ON snapshot (created_at)`,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
