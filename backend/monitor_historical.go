package main

import "time"

type MonitorHistorical struct {
	MonitorID         string        `json:"monitor_id"`
	Status            MonitorStatus `json:"status"`
	Latency           int64         `json:"latency"`
	Timestamp         time.Time     `json:"timestamp"`
	AdditionalMessage string        `json:"additional_message,omitempty"`
	HttpProtocol      string        `json:"http_protocol,omitempty"`
	TLSVersion        string        `json:"tls_version,omitempty"`
	TLSCipherName     string        `json:"tls_cipher_name,omitempty"`
	TLSExpiryDate     time.Time     `json:"tls_expiry_date,omitempty"`
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
