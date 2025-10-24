package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// KeyConfig represents a single virtual key configuration
type KeyConfig struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
}

// Config represents the keys.json structure
type Config struct {
	VirtualKeys map[string]KeyConfig `json:"virtual_keys"`
}

// Load reads and parses the configuration file
func Load(filepath string) (*Config, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &config, nil
}
