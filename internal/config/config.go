package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// RuleConfig represents the configuration settings for an individual rule.
type RuleConfig struct {
	Enabled  bool           `yaml:"enabled"`
	Severity string         `yaml:"severity,omitempty"`
	Params   map[string]any `yaml:",inline"`
}

// Config represents the root configuration structure loaded from archguard.yaml.
type Config struct {
	Version string                `yaml:"version"`
	Rules   map[string]RuleConfig `yaml:"rules"`
}

// Load reads and parses the YAML configuration file from the specified file path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file from %s: %w", path, err)
	}

	cfg := &Config{
		Rules: make(map[string]RuleConfig),
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml config from %s: %w", path, err)
	}

	return cfg, nil
}
