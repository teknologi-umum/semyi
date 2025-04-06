package main_test

import (
	"context"
	"testing"
	"time"

	main "semyi"
	"semyi/testutils"
)

// MockAlerter implements the Alerter interface for testing
type MockAlerter struct {
	alertsSent []main.AlertMessage
}

func (m *MockAlerter) Send(ctx context.Context, alert main.AlertMessage) error {
	m.alertsSent = append(m.alertsSent, alert)
	return nil
}

func TestProcessor_ProcessResponse(t *testing.T) {
	// Create mock dependencies
	mockWriter := main.NewMonitorHistoricalWriter(database)
	mockReader := main.NewMonitorHistoricalReader(database)
	mockBroker := main.NewBroker[main.MonitorHistorical]()
	mockAlerter := &MockAlerter{}

	// Create processor with mock dependencies
	processor := &main.Processor{
		HistoricalWriter:      mockWriter,
		HistoricalReader:      mockReader,
		CentralBroker:         mockBroker,
		TelegramAlertProvider: mockAlerter,
		DiscordAlertProvider:  mockAlerter,
		HTTPAlertProvider:     mockAlerter,
		SlackAlertProvider:    mockAlerter,
	}

	// Test cases
	tests := []struct {
		name           string
		response       main.Response
		expectedStatus main.MonitorStatus
	}{
		{
			name: "Successful Response",
			response: main.Response{
				Success: true,
				Monitor: main.Monitor{
					UniqueID: "test-monitor-1",
					Name:     "Test Monitor",
				},
				RequestDuration: int64(100 * time.Millisecond),
				Timestamp:       time.Now(),
			},
			expectedStatus: main.MonitorStatusSuccess,
		},
		{
			name: "Failed Response",
			response: main.Response{
				Success: false,
				Monitor: main.Monitor{
					UniqueID: "test-monitor-2",
					Name:     "Test Monitor",
				},
				RequestDuration: int64(200 * time.Millisecond),
				Timestamp:       time.Now(),
			},
			expectedStatus: main.MonitorStatusFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Process the response
			processor.ProcessResponse(context.Background(), tt.response)

			// Verify that the alert was sent
			testutils.AssertGreater(t, 0, len(mockAlerter.alertsSent), "Expected at least one alert to be sent")

			// Verify the alert content
			lastAlert := mockAlerter.alertsSent[len(mockAlerter.alertsSent)-1]
			testutils.AssertEqual(t, tt.response.Monitor.UniqueID, lastAlert.MonitorID, "Monitor ID should match")
			testutils.AssertEqual(t, tt.response.Success, lastAlert.Success, "Success status should match")
		})
	}
}

func TestProcessor_ProcessResponse_WithLongID(t *testing.T) {
	// Create mock dependencies
	mockWriter := main.NewMonitorHistoricalWriter(database)
	mockReader := main.NewMonitorHistoricalReader(database)
	mockBroker := main.NewBroker[main.MonitorHistorical]()
	mockAlerter := &MockAlerter{}

	// Create processor with mock dependencies
	processor := &main.Processor{
		HistoricalWriter:      mockWriter,
		HistoricalReader:      mockReader,
		CentralBroker:         mockBroker,
		TelegramAlertProvider: mockAlerter,
		DiscordAlertProvider:  mockAlerter,
		HTTPAlertProvider:     mockAlerter,
		SlackAlertProvider:    mockAlerter,
	}

	// Create a response with a very long ID
	longID := "a" + string(make([]byte, 300)) // Create a string longer than 255 characters
	response := main.Response{
		Success: true,
		Monitor: main.Monitor{
			UniqueID: longID,
			Name:     "Test Monitor",
		},
		RequestDuration: int64(100 * time.Millisecond),
		Timestamp:       time.Now(),
	}

	// Process the response
	processor.ProcessResponse(context.Background(), response)

	// Verify that the ID was truncated
	testutils.AssertLessOrEqual(t, 255, len(response.Monitor.UniqueID), "ID should be truncated to 255 characters")
}

func TestProcessor_ProcessResponse_WithError(t *testing.T) {
	// Create mock dependencies
	mockWriter := main.NewMonitorHistoricalWriter(database)
	mockReader := main.NewMonitorHistoricalReader(database)
	mockBroker := main.NewBroker[main.MonitorHistorical]()
	mockAlerter := &MockAlerter{}

	// Create processor with mock dependencies
	processor := &main.Processor{
		HistoricalWriter:      mockWriter,
		HistoricalReader:      mockReader,
		CentralBroker:         mockBroker,
		TelegramAlertProvider: mockAlerter,
		DiscordAlertProvider:  mockAlerter,
		HTTPAlertProvider:     mockAlerter,
		SlackAlertProvider:    mockAlerter,
	}

	// Create a response with an error
	response := main.Response{
		Success: false,
		Monitor: main.Monitor{
			UniqueID: "test-monitor-3",
			Name:     "Test Monitor",
		},
		RequestDuration: int64(100 * time.Millisecond),
		Timestamp:       time.Now(),
		StatusCode:      500,
	}

	// Process the response
	processor.ProcessResponse(context.Background(), response)

	// Verify that the alert was sent
	testutils.AssertGreater(t, 0, len(mockAlerter.alertsSent), "Expected at least one alert to be sent")
	lastAlert := mockAlerter.alertsSent[len(mockAlerter.alertsSent)-1]
	testutils.AssertEqual(t, response.StatusCode, lastAlert.StatusCode, "Status code should be included in the alert")
}
