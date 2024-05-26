package main_test

import (
	main "semyi"
	"testing"
)

func TestIncidentValidate(t *testing.T) {
	validPayload := main.Incident{
		MonitorID:   "a84c2c59-748c-48d0-b628-4a73b1c3a8d7",
		Title:       "test",
		Description: "description test",
		Timestamp:   "2024-05-26T15:04:05+07:00",
		Severity:    main.IncidentSeverityError,
		Status:      main.IncidentStatusInvestigating,
	}

	t.Run("Should return error if payload is invalid", func(t *testing.T) {
		t.Run("Timestamp", func(t *testing.T) {
			validPayloadCopy := validPayload
			mockTimestamps := []string{
				"2024-05-26T15:04:05",
				"2024-05-26",
				"15:04:05",
				"arbitary",
				"2024-05-26 15:04:05",
			}

			for _, timestamp := range mockTimestamps {
				validPayloadCopy.Timestamp = timestamp

				err := validPayloadCopy.Validate()
				if err == nil {
					t.Error("expect error, got nil")
				}
			}
		})
		t.Run("severity", func(t *testing.T) {
			validPayloadCopy := validPayload
			mockSeverity := []uint{4, 5, 6}

			for _, severity := range mockSeverity {
				validPayloadCopy.Severity = main.IncidentSeverity(severity)

				err := validPayloadCopy.Validate()
				if err == nil {
					t.Error("expect error, got nil")
				}
			}
		})
		t.Run("status", func(t *testing.T) {
			validPayloadCopy := validPayload
			mockStatus := []uint{5, 6, 7}

			for _, status := range mockStatus {
				validPayloadCopy.Status = main.IncidentStatus(status)

				err := validPayloadCopy.Validate()
				if err == nil {
					t.Error("expect error, got nil")
				}
			}
		})
	})

	t.Run("Shouldn't return error if payload is valid", func(t *testing.T) {
		err := validPayload.Validate()
		if err != nil {
			t.Errorf("expect error nil, but got %v", err)
		}
	})
}
