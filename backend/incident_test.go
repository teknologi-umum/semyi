package main_test

import (
	"errors"
	main "semyi"
	"testing"
	"time"
)

func TestIncidentValidate(t *testing.T) {
	validPayload := main.Incident{
		MonitorID:   "a84c2c59-748c-48d0-b628-4a73b1c3a8d7",
		Title:       "test",
		Description: "description test",
		Timestamp:   time.Date(2000, 7, 24, 4, 30, 15, 0, time.UTC),
		Severity:    main.IncidentSeverityError,
		Status:      main.IncidentStatusInvestigating,
	}

	t.Run("Should return error if payload is invalid", func(t *testing.T) {
		t.Run("Timestamp", func(t *testing.T) {
			validPayloadCopy := validPayload
			mockTimestamps := []time.Time{
				time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
			}

			for _, timestamp := range mockTimestamps {
				validPayloadCopy.Timestamp = timestamp

				err := validPayloadCopy.Validate()
				if err == nil {
					t.Error("expect error, got nil")
				}

				var expectError *main.ValidationError
				if !errors.As(err, &expectError) {
					t.Errorf("expect error: %T, but got : %T", expectError, err)
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

				var expectError *main.ValidationError
				if !errors.As(err, &expectError) {
					t.Errorf("expect error: %T, but got : %T", expectError, err)
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

				var expectError *main.ValidationError
				if !errors.As(err, &expectError) {
					t.Errorf("expect error: %T, but got : %T", expectError, err)
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
