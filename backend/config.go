package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
)

type Endpoint struct {
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Description string            `json:"description"`
	Timeout     int               `json:"timeout"`
	Interval    int               `json:"interval"`
	Headers     map[string]string `json:"headers"`
	Method      string            `json:"method"`
}

func ReadConfigurationFile(path string) ([]Endpoint, error) {
	if path == "" {
		path = "../config.json"
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open configuration file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %v", err)
	}

	var endpoints []Endpoint
	err = json.Unmarshal(data, &endpoints)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %v", err)
	}

	return endpoints, nil
}

func ValidateConfiguration(config Endpoint) (bool, error) {
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
