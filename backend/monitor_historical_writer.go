package main

import (
	"database/sql"
	"time"
)

type MonitorStatus uint8

const (
	MonitorStatusSuccess MonitorStatus = iota
	MonitorStatusFailure
)

type MonitorHistorical struct {
	MonitorID string
	Status    MonitorStatus
	Latency   int64
	Timestamp time.Time
}

type MonitorHistoricalWriter struct {
	db *sql.DB
}

func NewMonitorHistoricalWriter(db *sql.DB) *MonitorHistoricalWriter {
	return &MonitorHistoricalWriter{db: db}
}

func (w *MonitorHistoricalWriter) Write(historical MonitorHistorical) error {
	// TODO: Work on me
	return nil
}
