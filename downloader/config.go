package downloader

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the download manager configuration.
type Config struct {
	MaxConcurrentDownloads int    `json:"maxConcurrentDownloads"`
	MaxPartsPerDownload    int    `json:"maxPartsPerDownload"`
	DownloadDir            string `json:"downloadDir"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	return Config{
		MaxConcurrentDownloads: 3,
		MaxPartsPerDownload:    4,
		DownloadDir:            filepath.Join(home, "Downloads"),
	}
}

const configFileName = "godm_config.json"

// configPath returns the path to the persisted config file.
func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "godm", configFileName)
}

// LoadConfig reads the config from disk, returning defaults if not found.
func LoadConfig() Config {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return DefaultConfig()
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}
	return cfg
}

// Save persists the config to disk.
func (c Config) Save() error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
