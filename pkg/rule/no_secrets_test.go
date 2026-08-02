package rule

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/archguard/archguard/internal/core"
)

func TestNoSecretsRule_CleanFile(t *testing.T) {
	tempDir := t.TempDir()
	cleanFile := filepath.Join(tempDir, "config.go")
	if err := os.WriteFile(cleanFile, []byte(`package main; var dbHost = "localhost"`), 0644); err != nil {
		t.Fatalf("failed creating temp file: %v", err)
	}

	r := NewNoSecretsRule(core.SeverityError)
	ctx := core.NewScanContext(context.Background(), tempDir, []string{"config.go"})

	issues, err := r.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected 0 issues for clean code, got %d", len(issues))
	}
}

func TestNoSecretsRule_HardcodedSecrets(t *testing.T) {
	tempDir := t.TempDir()
	dirtyFile := filepath.Join(tempDir, "keys.go")
	secretCode := `package main
var awsKey = "AKIAIOSFODNN7EXAMPLE"
`
	if err := os.WriteFile(dirtyFile, []byte(secretCode), 0644); err != nil {
		t.Fatalf("failed creating temp file: %v", err)
	}

	r := NewNoSecretsRule(core.SeverityError)
	ctx := core.NewScanContext(context.Background(), tempDir, []string{"keys.go"})

	issues, err := r.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) == 0 {
		t.Errorf("expected issues for hardcoded AWS key, got 0")
	}
}
