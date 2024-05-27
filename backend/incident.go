package main

import (
	"time"
)

type IncidentSeverity uint

const (
	IncidentSeverityInformational IncidentSeverity = iota
	IncidentSeverityWarning
	IncidentSeverityError
	IncidentSeverityFatal
)

func (s IncidentSeverity) IsValid() bool {
	switch s {
	case IncidentSeverityInformational, IncidentSeverityWarning, IncidentSeverityError, IncidentSeverityFatal:
		return true
	}
	return false
}

type IncidentStatus uint

const (
	IncidentStatusInvestigating IncidentStatus = iota
	IncidentStatusIdentified
	IncidentStatusMonitoring
	IncidentStatusResolved
	IncidentStatusScheduled
)

func (s IncidentStatus) IsValid() bool {
	switch s {
	case IncidentStatusInvestigating, IncidentStatusIdentified, IncidentStatusMonitoring, IncidentStatusResolved, IncidentStatusScheduled:
		return true
	}
	return false
}

type Incident struct {
	MonitorID   string           `json:"monitor_id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Timestamp   time.Time        `json:"timestamp"`
	Severity    IncidentSeverity `json:"severity"`
	Status      IncidentStatus   `json:"status"`
	CreatedBy   string           `json:"created_by"`
}

func (i Incident) Validate() error {
	err := NewValidationError()

	if i.Timestamp.IsZero() {
		err.AddIssue("timestamp", "shouldn't be zero")
	}

	if !i.Severity.IsValid() {
		err.AddIssue("severity", "invalid")
	}

	if !i.Status.IsValid() {
		err.AddIssue("status", "invalid")
	}

	if err.HasIssues() {
		return err
	}

	return nil
}
