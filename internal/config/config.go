package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the application's configuration.
type Config struct {
	SavePath   string `json:"save_path"`
	BackupDir  string `json:"backup_dir"`
	AutoBackup bool   `json:"auto_backup"`
}

// Load loads the configuration from a file. If the file doesn't exist,
// it returns a default configuration and a 'first run' flag.
func Load() (*Config, bool, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, false, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{AutoBackup: true}, true, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, false, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, false, err
	}

	return &cfg, false, nil
}

// Save saves the configuration to a file.
func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// getConfigPath returns the path to the configuration file.
func getConfigPath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exePath), "config.json"), nil
}
