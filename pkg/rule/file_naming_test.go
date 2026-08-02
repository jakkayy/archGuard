package rule_test

import (
	"testing"

	"github.com/archguard/archguard/internal/core"
	"github.com/archguard/archguard/pkg/rule"
)

func TestFileNamingRule_ValidFiles(t *testing.T) {
	r, err := rule.NewFileNamingRule(`^[a-z0-9._-]+$`, core.SeverityWarning)
	if err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}

	ctx := core.NewScanContext(nil, ".", []string{
		"main.go",
		"file_naming.go",
		"config-file.yaml",
	})

	issues, err := r.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("expected 0 issues for valid filenames, got %d", len(issues))
	}
}

func TestFileNamingRule_InvalidFiles(t *testing.T) {
	r, err := rule.NewFileNamingRule(`^[a-z0-9._-]+$`, core.SeverityWarning)
	if err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}

	ctx := core.NewScanContext(nil, ".", []string{
		"InvalidFileName.go",
		"bad file name.ts",
	})

	issues, err := r.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 2 {
		t.Errorf("expected 2 issues for invalid filenames, got %d", len(issues))
	}

	if issues[0].Severity != core.SeverityWarning {
		t.Errorf("expected severity WARNING, got %s", issues[0].Severity)
	}
}
