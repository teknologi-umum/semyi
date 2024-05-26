package main

import (
	"fmt"
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
	Timestamp   string           `json:"timestamp"`
	Severity    IncidentSeverity `json:"severity"`
	Status      IncidentStatus   `json:"status"`
	CreatedBy   string           `json:"created_by"`
}

func (i Incident) Validate() error {
	_, err := time.Parse(time.RFC3339, i.Timestamp)
	if err != nil {
		return err
	}

	if !i.Severity.IsValid() {
		return fmt.Errorf("invalid incident severity")
	}

	if !i.Status.IsValid() {
		return fmt.Errorf("invalid incident status")
	}

	return nil
}
