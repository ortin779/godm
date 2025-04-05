package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	ConfigFilename      = "config.json"
	BaseDir             = ".godm"
	DefaultDownloadDir  = "Downloads"
	DefaultParts        = 5
	DefaultMaxDownloads = 100
)

type Config struct {
	DownloadDir  string `json:"download_dir"`
	MaxDownloads int    `json:"max_downloads"`
	DefaultParts int    `json:"default_parts"`
}

func GetConfig() (*Config, error) {
	cfg, err := loadConfig()

	return cfg, err
}

func UpdateConfig(downloadDir string, maxDownloads, parts int) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if cfg.DefaultParts != parts && parts > 0 {
		cfg.DefaultParts = parts
	}
	if cfg.DownloadDir != downloadDir && downloadDir != "" {
		cfg.DownloadDir = downloadDir
	}
	if cfg.MaxDownloads != maxDownloads && maxDownloads > 0 {
		cfg.MaxDownloads = maxDownloads
	}

	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configFilePath := filepath.Join(homeDir, BaseDir, ConfigFilename)
	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteAt(cfgBytes, 0)
	if err != nil {
		return err
	}
	return nil
}

func loadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	baseDirPath := filepath.Join(homeDir, BaseDir)
	if _, err := os.Stat(baseDirPath); os.IsNotExist(err) {
		err := os.Mkdir(baseDirPath, 0644)
		if err != nil {
			return nil, err
		}
	}

	configFilePath := filepath.Join(baseDirPath, ConfigFilename)
	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := new(Config)
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		defaultConfig := Config{
			DownloadDir:  filepath.Join(homeDir, DefaultDownloadDir),
			MaxDownloads: DefaultMaxDownloads,
			DefaultParts: DefaultParts,
		}

		cfgBytes, err := json.Marshal(defaultConfig)
		if err != nil {
			return nil, err
		}
		_, err = file.WriteAt(cfgBytes, 0)
		if err != nil {
			return nil, err
		}
		return &defaultConfig, err
	}

	return config, nil
}
