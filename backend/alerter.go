package main

import (
	"context"
	"time"
)

type Alerter interface {
	Send(ctx context.Context, msg AlertMessage) error
}

type AlertMessage struct {
	Success     bool
	StatusCode  int
	Timestamp   time.Time
	MonitorID   string
	MonitorName string
	Latency     int64
}
