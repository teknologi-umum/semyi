package main

import (
	"time"
)

// HttpCommonError represents a standardized error response
type HttpCommonError struct {
	Error string `json:"error"`
}

// HttpCommonSuccess represents a standardized success response
type HttpCommonSuccess struct {
	Message string `json:"message"`
}

// HttpCommonData represents a standardized data response
type HttpCommonData struct {
	Data interface{} `json:"data"`
}

// StaticSnapshotResponse represents the response for /api/static endpoint
type StaticSnapshotResponse struct {
	Metadata   Monitor             `json:"metadata"`
	Historical []MonitorHistorical `json:"historical"`
}

// MonitorHistoricalResponse represents a single monitor historical data point
type MonitorHistoricalResponse struct {
	MonitorID string        `json:"monitor_id"`
	Status    MonitorStatus `json:"status"`
	Latency   int64         `json:"latency"`
	Timestamp time.Time     `json:"timestamp"`
}
