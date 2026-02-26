package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ServiceConfig struct {
	APIKey string `json:"api_key"`
	UID    string `json:"uid,omitempty"`
}

type Config struct {
	Gemini *ServiceConfig `json:"gemini,omitempty"`
	Veo3   *ServiceConfig `json:"veo3,omitempty"`
	Ark     *ServiceConfig `json:"ark,omitempty"`
	TopView *ServiceConfig `json:"topview,omitempty"`
}

func Path() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "llm-api-plugin", "config.json")
}

func Load() (*Config, error) {
	path := Path()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config %s: %w\nRun '<cli> config set-key <KEY>' to configure", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config %s: %w", path, err)
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	path := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func LoadOrCreate() (*Config, error) {
	cfg, err := Load()
	if err != nil {
		return &Config{}, nil
	}
	return cfg, nil
}

// ResolveAPIKey returns the API key for a service.
// Priority: environment variable > config file.
func ResolveAPIKey(envVar string, fromConfig *ServiceConfig) string {
	if v := os.Getenv(envVar); v != "" {
		return v
	}
	if fromConfig != nil && fromConfig.APIKey != "" {
		return fromConfig.APIKey
	}
	return ""
}
