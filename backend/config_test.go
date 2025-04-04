package main_test

import (
	"os"
	"path/filepath"
	"testing"

	main "semyi"
)

func TestReadConfigurationFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		fileContent   string
		fileExtension string
		wantErr       bool
		errContains   string
	}{
		{
			name: "valid JSON configuration",
			fileContent: `{
				"monitors": [
					{
						"unique_id": "test-1",
						"name": "Test Monitor",
						"description": "Test Description",
						"type": "http",
						"interval": 60,
						"timeout": 30,
						"http_endpoint": "https://example.com"
					}
				],
				"webhook": {
					"monitorIds": "https://webhook.example.com",
					"success_response": true,
					"failed_response": true
				},
				"retention_period": 90
			}`,
			fileExtension: ".json",
			wantErr:       false,
		},
		{
			name: "valid YAML configuration",
			fileContent: `monitors:
  - unique_id: test-1
    name: Test Monitor
    description: Test Description
    type: http
    interval: 60
    timeout: 30
    http_endpoint: https://example.com
webhook:
  monitorIds: https://webhook.example.com
  success_response: true
  failed_response: true
retention_period: 90`,
			fileExtension: ".yaml",
			wantErr:       false,
		},
		{
			name: "valid TOML configuration",
			fileContent: `[[monitors]]
unique_id = "test-1"
name = "Test Monitor"
description = "Test Description"
type = "http"
interval = 60
timeout = 30
http_endpoint = "https://example.com"

[webhook]
monitorIds = "https://webhook.example.com"
success_response = true
failed_response = true

retention_period = 90`,
			fileExtension: ".toml",
			wantErr:       false,
		},
		{
			name:          "invalid file format",
			fileContent:   "invalid content",
			fileExtension: ".txt",
			wantErr:       true,
			errContains:   "invalid configuration file format",
		},
		{
			name:          "malformed JSON",
			fileContent:   `{"invalid": json}`,
			fileExtension: ".json",
			wantErr:       true,
			errContains:   "failed to parse configuration file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file
			tempFile := filepath.Join(tempDir, "config"+tt.fileExtension)
			err := os.WriteFile(tempFile, []byte(tt.fileContent), 0644)
			if err != nil {
				t.Fatalf("failed to create temporary file: %v", err)
			}

			// Clean up the temporary file after the test
			t.Cleanup(func() {
				if err := os.Remove(tempFile); err != nil {
					t.Logf("failed to remove temporary file %s: %v", tempFile, err)
				}
			})

			// Test the function
			config, err := main.ReadConfigurationFile(tempFile)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("error message %q does not contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Validate the configuration
			if len(config.Monitors) == 0 {
				t.Error("expected at least one monitor in configuration")
			}

			// Check if defaults are set
			if config.RetentionPeriod <= 0 {
				t.Error("retention period should be set to a positive value")
			}
		})
	}

	// Test non-existent file
	t.Run("non-existent file", func(t *testing.T) {
		nonexistentFile := filepath.Join(tempDir, "nonexistent.json")
		_, err := main.ReadConfigurationFile(nonexistentFile)
		if err == nil {
			t.Error("expected error for non-existent file but got none")
		}
	})
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
