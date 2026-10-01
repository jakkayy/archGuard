package rule

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jakkayy/archGuard/internal/core"
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
	secretCode := "package main\nvar awsKey = \"AKIA" + "IOSFODNN7EXAMPLE\"\n"
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

func runNoSecrets(t *testing.T, name, content string) []core.Issue {
	t.Helper()
	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, name), []byte(content), 0644); err != nil {
		t.Fatalf("failed creating temp file: %v", err)
	}
	r := NewNoSecretsRule(core.SeverityError)
	issues, err := r.Run(core.NewScanContext(context.Background(), tempDir, []string{name}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return issues
}

func TestNoSecretsRule_ReportsLineNumberPerOccurrence(t *testing.T) {
	// Fixtures are split with "+" so ArchGuard's own self-scan does not flag this file.
	issues := runNoSecrets(t, "keys.go", "package main\n"+
		"var b = \"AKIA"+"IOSFODNN7EXAMPL2\"\n"+
		"\n"+
		"var c = \"AKIA"+"IOSFODNN7EXAMPL3\"\n")

	if len(issues) != 2 {
		t.Fatalf("expected 2 issues, got %d: %+v", len(issues), issues)
	}
	if issues[0].Line != 2 || issues[1].Line != 4 {
		t.Errorf("expected lines 2 and 4, got %d and %d", issues[0].Line, issues[1].Line)
	}
}

func TestNoSecretsRule_DetectsRealPrivateKeyHeaders(t *testing.T) {
	for _, header := range []string{
		"RSA PRIVATE KEY",
		"OPENSSH PRIVATE KEY",
		"EC PRIVATE KEY",
		"ENCRYPTED PRIVATE KEY",
		"PRIVATE KEY",
	} {
		issues := runNoSecrets(t, "id_key", "-----BEGIN "+header+"-----\nMIIEow...\n")
		if len(issues) != 1 {
			t.Errorf("expected private key header %q to be detected, got %d issues", header, len(issues))
		}
	}
}

func TestNoSecretsRule_IgnoreDirectiveSuppressesLine(t *testing.T) {
	issues := runNoSecrets(t, "fixture.go", "var k = \"AKIA"+"IOSFODNN7EXAMPLE\" // archguard:ignore\n")
	if len(issues) != 0 {
		t.Errorf("expected ignore directive to suppress finding, got %d issues", len(issues))
	}
}

func TestNoSecretsRule_SkipsBinaryFiles(t *testing.T) {
	issues := runNoSecrets(t, "blob.bin", "\x00\x01AKIA"+"IOSFODNN7EXAMPLE")
	if len(issues) != 0 {
		t.Errorf("expected binary file to be skipped, got %d issues", len(issues))
	}
}

func TestNoSecretsRule_DetectsGitHubToken(t *testing.T) {
	issues := runNoSecrets(t, "ci.sh", "TOKEN=ghp_"+strings.Repeat("a", 36)+"\n")
	if len(issues) != 1 {
		t.Errorf("expected GitHub token to be detected, got %d issues", len(issues))
	}
}
