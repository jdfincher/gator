// Package config implements basic user/database config from .gatorconfig.json
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DBURL    string `json:"db_url"`
	UserName string `json:"current_user_name"`
}

func Read() (*Config, error) {
	config := new(Config)
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return config, err
	}
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return config, fmt.Errorf("error reading file -> %w \nwith path -> %v", err, configFilePath)
	}
	if err := json.Unmarshal(data, config); err != nil {
		return config, fmt.Errorf("error decoding config -> %w", err)
	}
	return config, nil
}

func (c *Config) SetUser(name string) error {
	c.UserName = name
	if err := write(c); err != nil {
		return err
	}
	return nil
}

func getConfigFilePath() (string, error) {
	const configFileName = ".gatorconfig.json"
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting config filepath -> %w", err)
	}
	filepath := home + "/" + configFileName
	return filepath, nil
}

func write(cfg *Config) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error encoding data for writing -> %w", err)
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing file -> %w", err)
	}
	return nil
}
