package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIHost  string `json:"api_host"`
	Token    string `json:"token"`
	APIToken string `json:"api_token"`
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".paas-cli.json")
}

func Load() *Config {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return &Config{APIHost: "https://api.espace-tech.com"}
	}

	var cfg Config
	json.Unmarshal(data, &cfg)
	if cfg.APIHost == "" {
		cfg.APIHost = "https://api.espace-tech.com"
	}
	return &cfg
}

func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0600)
}