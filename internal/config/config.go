package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// RuleConfig represents the configuration settings for an individual rule.
type RuleConfig struct {
	Enabled  bool           `yaml:"enabled"`
	Severity string         `yaml:"severity,omitempty"`
	Params   map[string]any `yaml:",inline"`
}

// UnmarshalYAML implements custom YAML unmarshaling for RuleConfig to capture extra rule parameters.
func (r *RuleConfig) UnmarshalYAML(value *yaml.Node) error {
	type rawConfig struct {
		Enabled  bool   `yaml:"enabled"`
		Severity string `yaml:"severity"`
	}

	var raw rawConfig
	if err := value.Decode(&raw); err != nil {
		return err
	}

	r.Enabled = raw.Enabled
	r.Severity = raw.Severity

	var fullMap map[string]any
	if err := value.Decode(&fullMap); err != nil {
		return err
	}

	delete(fullMap, "enabled")
	delete(fullMap, "severity")
	r.Params = fullMap

	return nil
}

// Config represents the root configuration structure loaded from archguard.yaml.
type Config struct {
	Version string                `yaml:"version"`
	Ignore  []string              `yaml:"ignore,omitempty"`
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

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config %s: %w", path, err)
	}

	return cfg, nil
}

// SupportedVersion is the only configuration schema version understood by this release.
const SupportedVersion = "v1"

// Validate checks the schema version and normalizes rule severities in place.
func (c *Config) Validate() error {
	if c.Version == "" {
		c.Version = SupportedVersion
	}
	if c.Version != SupportedVersion {
		return fmt.Errorf("unsupported config version %q (expected %q)", c.Version, SupportedVersion)
	}

	ids := make([]string, 0, len(c.Rules))
	for id := range c.Rules {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		ruleCfg := c.Rules[id]
		sev, err := NormalizeSeverity(ruleCfg.Severity)
		if err != nil {
			return fmt.Errorf("rule %q: %w", id, err)
		}
		ruleCfg.Severity = sev
		c.Rules[id] = ruleCfg
	}

	return nil
}

// NormalizeSeverity converts a user-supplied severity (case-insensitive, "warn" allowed)
// into its canonical upper-case form. An empty value is returned unchanged so rules
// can apply their own default.
func NormalizeSeverity(s string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "":
		return "", nil
	case "ERROR":
		return "ERROR", nil
	case "WARNING", "WARN":
		return "WARNING", nil
	case "INFO":
		return "INFO", nil
	default:
		return "", fmt.Errorf("invalid severity %q (allowed: ERROR, WARNING, INFO)", s)
	}
}
