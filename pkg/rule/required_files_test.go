package rule

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/archguard/archguard/internal/core"
)

func TestRequiredFilesRule_FileExists(t *testing.T) {
	tempDir := t.TempDir()
	readmePath := filepath.Join(tempDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed creating temp file: %v", err)
	}

	r := NewRequiredFilesRule([]string{"README.md"}, core.SeverityError)
	ctx := core.NewScanContext(context.Background(), tempDir, []string{"README.md"})

	issues, err := r.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}

func TestRequiredFilesRule_FileMissing(t *testing.T) {
	tempDir := t.TempDir()

	r := NewRequiredFilesRule([]string{"README.md", "LICENSE"}, core.SeverityError)
	ctx := core.NewScanContext(context.Background(), tempDir, []string{})

	issues, err := r.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 2 {
		t.Errorf("expected 2 issues for missing files, got %d", len(issues))
	}
}
