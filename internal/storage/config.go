package storage

import (
	"encoding/json"
	"os"
)

type Config struct {
	FrequencyHours int `json:"frequency_hours"`
}

const configFile = "gospeed_config.json"

// LoadConfig loads the configuration from file, returns default if file doesn't exist
func LoadConfig() (Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config (no automatic tests)
			return Config{FrequencyHours: 0}, nil
		}
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

// SaveConfig saves the configuration to file
func SaveConfig(config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}
