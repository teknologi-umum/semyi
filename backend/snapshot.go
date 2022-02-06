package main

import (
	"context"
	"database/sql"
)

func (d *Deps) WriteSnapshot(ctx context.Context, items []Response) error {
	c, err := d.DB.Conn(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	tx, err := c.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadUncommitted, ReadOnly: false})
	if err != nil {
		return err
	}

	for _, item := range items {
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO
				snapshot
			(
				url,
				timeout,
				interval,
				status_code,
				request_duration,
				created_at
			)
			VALUES
			(
				$1, $2, $3, $4, $5, $6
			)`,
			item.Endpoint.URL,
			item.Endpoint.Timeout,
			item.Endpoint.Interval,
			item.StatusCode,
			item.RequestDuration,
			item.Timestamp,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
