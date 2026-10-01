package rule_test

import (
	"context"
	"strings"
	"testing"

	"github.com/jakkayy/archGuard/internal/core"
	"github.com/jakkayy/archGuard/pkg/rule"
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

func TestFileNamingRule_SuggestionReflectsCustomPattern(t *testing.T) {
	r, err := rule.NewFileNamingRule(`^[A-Z][a-zA-Z]+\.tsx$`, core.SeverityWarning)
	if err != nil {
		t.Fatal(err)
	}
	issues, err := r.Run(core.NewScanContext(context.Background(), t.TempDir(), []string{"components/button.tsx"}))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if !strings.Contains(issues[0].Suggestion, `^[A-Z][a-zA-Z]+\.tsx$`) {
		t.Errorf("expected suggestion to mention the configured pattern, got %q", issues[0].Suggestion)
	}
}
