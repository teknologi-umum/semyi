package main

import (
	"testing"
	"time"
)

func TestEnsureUTC(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "UTC time should remain unchanged",
			input:    time.Date(2024, 5, 25, 12, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 5, 25, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Local time should be converted to UTC",
			input:    time.Date(2024, 5, 25, 12, 0, 0, 0, time.Local),
			expected: time.Date(2024, 5, 25, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "Fixed offset timezone should be converted to UTC",
			input:    time.Date(2024, 5, 25, 12, 0, 0, 0, time.FixedZone("UTC+7", 7*60*60)),
			expected: time.Date(2024, 5, 25, 5, 0, 0, 0, time.UTC),
		},
		{
			name:     "Negative offset timezone should be converted to UTC",
			input:    time.Date(2024, 5, 25, 12, 0, 0, 0, time.FixedZone("UTC-7", -7*60*60)),
			expected: time.Date(2024, 5, 25, 19, 0, 0, 0, time.UTC),
		},
		{
			name:     "Zero time should remain zero",
			input:    time.Time{},
			expected: time.Time{},
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnsureUTC(tt.input)

			// Check if the time is in UTC
			if got.Location() != time.UTC {
				t.Errorf("EnsureUTC() location = %v, want %v", got.Location(), time.UTC)
			}

			// Check if the time value is correct
			if !got.Equal(tt.expected) {
				t.Errorf("EnsureUTC() = %v, want %v", got, tt.expected)
			}
		})
	}
}
