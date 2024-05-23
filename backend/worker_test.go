package main_test

import (
	"strconv"
	"strings"
	"testing"
)

func parseExpectedStatusCode(expected string, got int) bool {
	// Valid values:
	// * 200 -> Direct 200 status code
	// * 2xx -> Any 2xx status code (200-299)
	// * 200-300 -> Any 200-300 status code (inclusive)
	// * 2xx-3xx -> Any 2xx (200-299) and 3xx (300-399) status code (inclusive)

	if expected == strconv.Itoa(got) {
		return true
	}

	parts := strings.Split(expected, "-")
	var ok = false
	for _, part := range parts {
		if ok == true {
			break
		}
		expectedSmallerParts := strings.Split(part, "")
		gotSmallerParts := strings.Split(strconv.Itoa(got), "")

		for i, expectedPart := range expectedSmallerParts {
			if expectedPart == "x" {
				continue
			}

			if expectedPart == gotSmallerParts[i] {
				ok = true
				continue
			}

			if expectedPart != gotSmallerParts[i] {
				ok = false
				break
			}
		}
	}

	return ok
}

func TestParseExpectedStatusCode(t *testing.T) {
	tests := []struct {
		expected string
		got      int
		want     bool
	}{
		{"200", 200, true},
		{"2xx", 200, true},
		{"2xx", 201, true},
		{"2xx", 299, true},
		{"2xx", 300, false},
		{"200-300", 200, true},
		{"200-300", 300, true},
		{"200-300", 301, false},
		{"2xx-3xx", 200, true},
		{"2xx-3xx", 201, true},
		{"2xx-3xx", 299, true},
		{"2xx-3xx", 300, true},
		{"2xx-3xx", 301, true},
		{"2xx-3xx", 399, true},
		{"2xx-3xx", 400, false},
		{"2xx-5xx", 200, true},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := parseExpectedStatusCode(tt.expected, tt.got)
			if got != tt.want {
				t.Errorf("parseExpectedStatusCode(%s, %d) = %v; want %v", tt.expected, tt.got, got, tt.want)
			}
		})
	}
}
