package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/archguard/archguard/internal/config"
	"github.com/archguard/archguard/internal/core"
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

	if ruleCfg.Severity != core.SeverityWarning {
		t.Errorf("expected severity WARNING, got: %s", ruleCfg.Severity)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("non_existent_file.yaml")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}
