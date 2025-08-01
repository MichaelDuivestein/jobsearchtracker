package config

import (
	"encoding/json"
	"errors"
	"os"
)

type Config struct {
	DatabaseFilePath                     string `json:"database file path"`
	DatabaseFileName                     string `json:"database file name"`
	IsDatabaseFileLocationAbsolutePath   bool   `json:"is database file location absolute path"`
	DatabaseMigrationsPath               string `json:"database migrations path"`
	IsDatabaseMigrationsPathAbsolutePath bool   `json:"is database migrations path absolute path"`
	ServerPort                           int    `json:"server port"`
}

func NewConfig() (*Config, error) {
	return loadConfigFromFile("configs/config.json")
}

func loadConfigFromFile(filePathAndName string) (*Config, error) {
	data, err := os.ReadFile(filePathAndName)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	err = config.validate()
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (config *Config) validate() error {
	if config.DatabaseFilePath == "" {
		return errors.New("config.DatabaseFilePath is empty")
	}

	if config.DatabaseFileName == "" {
		return errors.New("config.DatabaseFileName is empty")
	}
	if config.DatabaseMigrationsPath == "" {
		return errors.New("config.DatabaseMigrationsPath is empty")
	}

	if config.ServerPort <= 0 || config.ServerPort > 65535 {
		return errors.New("config.ServerPort is invalid")
	}

	return nil
}
