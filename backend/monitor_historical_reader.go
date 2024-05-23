package main

import (
	"context"
	"database/sql"
)

type MonitorHistoricalReader struct {
	db *sql.DB
}

func NewMonitorHistoricalReader(db *sql.DB) *MonitorHistoricalReader {
	return &MonitorHistoricalReader{db: db}
}

func (r *MonitorHistoricalReader) ReadRawHistorical(ctx context.Context, monitorId string) ([]MonitorHistorical, error) {
	panic("TODO: implement me!")
}

func (r *MonitorHistoricalReader) ReadHourlyHistorical(ctx context.Context, monitorId string) ([]MonitorHistorical, error) {
	panic("TODO: implement me!")
}

func (r *MonitorHistoricalReader) ReadDailyHistorical(ctx context.Context, monitorId string) ([]MonitorHistorical, error) {
	panic("TODO: implement me!")
}

func (r *MonitorHistoricalReader) ReadRawLatest(ctx context.Context, monitorId string) (MonitorHistorical, error) {
	// Get the latest entry from the raw historical table
	panic("TODO: implement me!")
}
