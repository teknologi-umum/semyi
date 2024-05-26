package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type ConfigurationFile struct {
	Monitors []Monitor `json:"monitors"`
	Webhook  Webhook   `json:"webhook"`
}

type MonitorType string

const (
	MonitorTypeHTTP MonitorType = "http"
	MonitorTypePing MonitorType = "ping"
)

type Monitor struct {
	// UniqueID specifies unique identifier for each monitor. In any case of the monitor configuration value get
	// changed (name, description, public monitorIds, etc), if users want to keep the data intact, they should keep the
	// UniqueID the same.
	UniqueID string `json:"unique_id" yaml:"unique_id" toml:"unique_id"`
	// Name specifies the display name that will be shown in the dashboard.
	Name string `json:"name" yaml:"name" toml:"name"`
	// Description specifies the description of the monitor. This is helpful as a friendly description of what
	// we are monitoring (e.g., "Push notification for email and SMS").
	Description string `json:"description" yaml:"description" toml:"description"`
	// PublicUrl specifies the public URL that will be shown in the dashboard. This is helpful to provide a different
	// public URL rather than providing the exact URL that's used for the HTTP monitor.
	PublicUrl string `json:"public_url" yaml:"public_url" toml:"public_url"`
	// Type specifies the type of monitor. It can be either "http" or "ping".
	Type MonitorType `json:"type" yaml:"type" toml:"type"`
	// Interval specifies the interval of each check in seconds. It must not be less or equal to zero.
	Interval int `json:"interval" yaml:"interval" toml:"interval"`
	// Timeout specifies the timeout for each check in seconds. It must not be less or equal to than zero.
	Timeout int `json:"timeout" yaml:"timeout" toml:"timeout"`
	// HttpHeaders specifies additional headers that are used for the HTTP request. It's a key-value pair where the key
	// specifies the header name and the value specifies the header value. This is optional.
	HttpHeaders map[string]string `json:"http_headers" yaml:"http_headers" toml:"http_headers"`
	// HttpMethod specifies the HTTP method that will be used for the HTTP request. It can be anything.
	// If not provided, it'll default to "GET".
	HttpMethod string `json:"http_method" yaml:"http_method" toml:"http_method"`
	// HttpEndpoint specifies the HTTP monitor that will be used for the HTTP request. It must be a valid URL.
	HttpEndpoint string `json:"http_endpoint" yaml:"http_endpoint" toml:"http_endpoint"`
	// HttpExpectedStatusCode specifies the expected status code for the HTTP request. If the status code is not the same
	// as the expected status code, it'll be considered as a failed check. The format of the value follows Caddy's health
	// check format: 200, 2xx, 200-300, 200-400, 2xx-4xx. This is optional. Defaults to 2xx.
	HttpExpectedStatusCode string `json:"http_expected_status_code" yaml:"http_expected_status_code" toml:"http_expected_status_code"`
	// IcmpHostname specifies the hostname that will be used for the ICMP request. It must be a valid hostname.
	IcmpHostname string `json:"hostname" yaml:"hostname" toml:"hostname"`
	// IcmpPacketSize specifies the packet size that will be used for the ICMP request. It must be greater than zero.
	// The default packet size is 56 bytes.
	IcmpPacketSize int `json:"packet_size" yaml:"packet_size" toml:"packet_size"`
}

func (m Monitor) MarshalJSON() ([]byte, error) {
	// We can't let everything be marshaled as is because we don't want to expose the configuration to be public.
	return json.Marshal(map[string]any{
		"id":          m.UniqueID,
		"name":        m.Name,
		"description": m.Description,
		"public_url":  m.PublicUrl,
	})
}

type Webhook struct {
	URL             string `json:"monitorIds" yaml:"monitorIds" toml:"monitorIds"`
	SuccessResponse bool   `json:"success_response" yaml:"success_response" toml:"success_response"`
	FailedResponse  bool   `json:"failed_response" yaml:"failed_response" toml:"failed_response"`
}

func ReadConfigurationFile(filePath string) (ConfigurationFile, error) {
	if filePath == "" {
		filePath = "../config.json"
	}

	file, err := os.Open(filePath)
	if err != nil {
		return ConfigurationFile{}, fmt.Errorf("failed to open configuration file: %w", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close configuration file")
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return ConfigurationFile{}, fmt.Errorf("failed to read configuration file: %w", err)
	}

	var configurationFile ConfigurationFile

	switch path.Ext(filePath) {
	case ".json":
		err := json.Unmarshal(data, &configurationFile)
		if err != nil {
			return ConfigurationFile{}, fmt.Errorf("failed to parse configuration file: %w", err)
		}
		break
	case ".yml":
		fallthrough
	case ".yaml":
		err := yaml.Unmarshal(data, &configurationFile)
		if err != nil {
			return ConfigurationFile{}, fmt.Errorf("failed to parse configuration file: %w", err)
		}
	case ".toml":
		err := toml.Unmarshal(data, &configurationFile)
		if err != nil {
			return ConfigurationFile{}, fmt.Errorf("failed to parse configuration file: %w", err)
		}
	default:
		return ConfigurationFile{}, fmt.Errorf("invalid configuration file format")
	}

	return configurationFile, nil
}

func (m Monitor) Validate() (bool, error) {
	if m.UniqueID == "" {
		return false, fmt.Errorf("unique_id is required")
	}

	if m.Name == "" {
		return false, fmt.Errorf("name is required")
	}

	if m.Timeout < 0 {
		return false, fmt.Errorf("timeout must be greater than 0")
	}

	if m.Interval < 0 {
		return false, fmt.Errorf("interval must be greater than 0")
	}

	switch m.Type {
	case MonitorTypeHTTP:
		if m.HttpEndpoint == "" {
			return false, fmt.Errorf("monitor is required")
		} else {
			// try to parse monitorIds
			_, err := url.Parse(m.HttpEndpoint)
			if err != nil {
				return false, fmt.Errorf("invalid monitorIds: %v", err)
			}
		}

	case MonitorTypePing:
		if m.IcmpHostname == "" {
			return false, fmt.Errorf("hostname is required")
		}
	default:
		return false, fmt.Errorf("invalid monitor type")
	}

	return true, nil
}

func ValidateWebhook(webhook Webhook) (bool, error) {
	if webhook.URL != "" {
		// Try to parse the given URL
		_, err := url.Parse(webhook.URL)
		if err != nil {
			return false, fmt.Errorf("invalid monitorIds: %v", err)
		}
	}

	if !webhook.FailedResponse && !webhook.SuccessResponse {
		return false, fmt.Errorf("failed_response and success_response cannot both be false")
	}

	return true, nil
}
