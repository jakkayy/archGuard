package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jakkayy/archGuard/internal/config"
)

func TestLoad_Success(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "archguard.yaml")

	content := `
version: "v1"
rules:
  file-naming:
    enabled: true
    severity: WARNING
    pattern: "^[a-z0-9_\\-\\.]+$"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.Version != "v1" {
		t.Errorf("expected version 'v1', got: %s", cfg.Version)
	}

	ruleCfg, ok := cfg.Rules["file-naming"]
	if !ok {
		t.Fatal("expected 'file-naming' rule in config")
	}

	if !ruleCfg.Enabled {
		t.Error("expected rule to be enabled")
	}

	if ruleCfg.Severity != "WARNING" {
		t.Errorf("expected severity WARNING, got: %s", ruleCfg.Severity)
	}

	pattern, ok := ruleCfg.Params["pattern"].(string)
	if !ok || pattern == "" {
		t.Errorf("expected pattern in Params, got: %v", ruleCfg.Params)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("non_existent_file.yaml")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "archguard.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}
	return path
}

func TestLoad_NormalizesSeverity(t *testing.T) {
	path := writeConfig(t, `
version: "v1"
rules:
  no-secrets:
    enabled: true
    severity: error
  file-naming:
    enabled: true
    severity: warn
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if got := cfg.Rules["no-secrets"].Severity; got != "ERROR" {
		t.Errorf("expected lowercase 'error' to normalize to ERROR, got %q", got)
	}
	if got := cfg.Rules["file-naming"].Severity; got != "WARNING" {
		t.Errorf("expected 'warn' to normalize to WARNING, got %q", got)
	}
}

func TestLoad_InvalidSeverity(t *testing.T) {
	path := writeConfig(t, `
rules:
  no-secrets:
    enabled: true
    severity: critical
`)
	if _, err := config.Load(path); err == nil {
		t.Fatal("expected error for invalid severity, got nil")
	}
}

func TestLoad_UnsupportedVersion(t *testing.T) {
	path := writeConfig(t, `version: "v2"`)
	if _, err := config.Load(path); err == nil {
		t.Fatal("expected error for unsupported version, got nil")
	}
}

func TestLoad_DefaultsVersion(t *testing.T) {
	path := writeConfig(t, `rules: {}`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Version != config.SupportedVersion {
		t.Errorf("expected default version %q, got %q", config.SupportedVersion, cfg.Version)
	}
}
