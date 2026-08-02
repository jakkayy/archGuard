package rule_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/archguard/archguard/internal/core"
	"github.com/archguard/archguard/pkg/rule"
)

func TestOpenAPIExistsRule_FileExists(t *testing.T) {
	tempDir := t.TempDir()
	specPath := "docs/openapi.json"
	fullPath := filepath.Join(tempDir, specPath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create temp spec file: %v", err)
	}

	r := rule.NewOpenAPIExistsRule(specPath, core.SeverityError)
	ctx := core.NewScanContext(nil, tempDir, []string{})

	issues, err := r.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("expected 0 issues when spec exists, got %d", len(issues))
	}
}

func TestOpenAPIExistsRule_FileMissing(t *testing.T) {
	tempDir := t.TempDir()
	specPath := "docs/non_existent.json"

	r := rule.NewOpenAPIExistsRule(specPath, core.SeverityError)
	ctx := core.NewScanContext(nil, tempDir, []string{})

	issues, err := r.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue when spec is missing, got %d", len(issues))
	}

	if issues[0].Severity != core.SeverityError {
		t.Errorf("expected severity ERROR, got %s", issues[0].Severity)
	}
}
