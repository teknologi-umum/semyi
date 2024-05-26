package main

import "time"

type MonitorHistorical struct {
	MonitorID string
	Status    MonitorStatus
	Latency   int64
	Timestamp time.Time
}

func (m MonitorHistorical) Validate() (bool, error) {
	validationError := NewValidationError()

	if m.MonitorID == "" {
		validationError.AddIssue("monitor_id", "monitor id is required")
	}

	if m.Latency < 0 {
		validationError.AddIssue("latency", "latency must be greater than 0")
	}

	if m.Timestamp.IsZero() {
		validationError.AddIssue("timestamp", "timestamp is required")
	}

	if m.Status != MonitorStatusSuccess && m.Status != MonitorStatusFailure {
		validationError.AddIssue("status", "invalid status")
	}

	if validationError.HasIssues() {
		return false, validationError
	}

	return true, nil
}
