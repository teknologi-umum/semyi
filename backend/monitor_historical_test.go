package main_test

import (
	"errors"
	"testing"
	"time"

	main "semyi"
)

func TestMonitorHistorical_Validate(t *testing.T) {
	t.Run("valid historical data", func(t *testing.T) {
		m := main.MonitorHistorical{
			MonitorID: "monitor-1",
			Status:    main.MonitorStatusSuccess,
			Latency:   123,
			Timestamp: time.Now(),
		}

		ok, err := m.Validate()
		if !ok {
			t.Errorf("expected historical data to be valid, got invalid")
		}

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("invalid historical data", func(t *testing.T) {
		m := main.MonitorHistorical{
			MonitorID: "",
			Status:    100,
			Latency:   -1,
			Timestamp: time.Time{},
		}

		ok, err := m.Validate()
		if ok {
			t.Errorf("expected historical data to be invalid, got valid")
		}

		if err == nil {
			t.Errorf("expected an error, got nil")
		}

		if err != nil {
			var validationError *main.ValidationError
			if errors.As(err, &validationError) {
				if !validationError.HasIssues() {
					t.Errorf("expected validation error to have issues, got none")
				}

				if len(validationError.Issues) != 4 {
					t.Errorf("expected 4 issues, got %d", len(validationError.Issues))
				}
			} else {
				t.Errorf("expected an error of type ValidationError, got %T", err)

			}
		}
	})
}
