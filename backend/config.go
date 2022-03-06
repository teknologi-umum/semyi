package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
)

type ConfigurationFile struct {
	Endpoints []Endpoint `json:"endpoints"`
	Webhook   Webhook    `json:"webhook"`
}

type Endpoint struct {
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Description string            `json:"description"`
	Timeout     int               `json:"timeout"`
	Interval    int               `json:"interval"`
	Headers     map[string]string `json:"headers"`
	Method      string            `json:"method"`
}

type Webhook struct {
	URL             string `json:"url"`
	SuccessResponse bool   `json:"success_response"`
	FailedResponse  bool   `json:"failed_response"`
}

func ReadConfigurationFile(path string) (ConfigurationFile, error) {
	if path == "" {
		path = "../config.json"
	}

	file, err := os.Open(path)
	if err != nil {
		return ConfigurationFile{}, fmt.Errorf("failed to open configuration file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return ConfigurationFile{}, fmt.Errorf("failed to read configuration file: %v", err)
	}

	var configurationFile ConfigurationFile
	err = json.Unmarshal(data, &configurationFile)
	if err != nil {
		return ConfigurationFile{}, fmt.Errorf("failed to parse configuration file: %v", err)
	}

	return configurationFile, nil
}

func ValidateEndpoint(config Endpoint) (bool, error) {
	if config.Name == "" {
		return false, fmt.Errorf("name is required")
	}

	if config.URL == "" {
		return false, fmt.Errorf("url is required")
	} else {
		// try to parse url
		_, err := url.Parse(config.URL)
		if err != nil {
			return false, fmt.Errorf("invalid url: %v", err)
		}
	}

	if config.Description == "" {
		return false, fmt.Errorf("description is required")
	}

	if config.Timeout < 0 {
		return false, fmt.Errorf("timeout must be greater than 0")
	}

	if config.Interval < 0 {
		return false, fmt.Errorf("interval must be greater than 0")
	}

	return true, nil
}

func ValidateWebhook(webhook Webhook) (bool, error) {
	if webhook.URL != "" {
		// Try to parse the given URL
		_, err := url.Parse(webhook.URL)
		if err != nil {
			return false, fmt.Errorf("invalid url: %v", err)
		}
	}

	if !webhook.FailedResponse && !webhook.SuccessResponse {
		return false, fmt.Errorf("failed_response and success_response cannot both be false")
	}

	return true, nil
}
