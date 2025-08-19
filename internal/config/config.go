package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure from .githooksrc.yml
type Config struct {
	Timeout  string                  `yaml:"timeout"`
	LogLevel string                  `yaml:"log_level"`
	Hooks    map[string][]HookCommand `yaml:"hooks"`
}

// HookCommand represents a single command to be executed for a hook.
type HookCommand struct {
	Run         string `yaml:"run"`
	Description string `yaml:"description"`
}

// Load reads the configuration file from the given path and returns a Config struct.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}
