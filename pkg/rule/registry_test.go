package rule_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jakkayy/archGuard/pkg/policy"
	"github.com/jakkayy/archGuard/pkg/rule"
)

func TestBuiltins_DefinitionsMatchBuiltRules(t *testing.T) {
	for _, d := range rule.Builtins().Definitions() {
		r, err := d.Build(nil, "")
		if err != nil {
			t.Fatalf("%s: Build with defaults failed: %v", d.ID, err)
		}
		if r.ID() != d.ID {
			t.Errorf("definition %q built rule with ID %q", d.ID, r.ID())
		}
		if r.Severity() != d.DefaultSeverity {
			t.Errorf("%s: default severity = %s, want %s", d.ID, r.Severity(), d.DefaultSeverity)
		}
	}
}

func TestDefinition_Build_RejectsUnknownParam(t *testing.T) {
	d, _ := rule.Builtins().Lookup("file-naming")
	_, err := d.Build(map[string]any{"patern": "^x$"}, "")
	if err == nil || !strings.Contains(err.Error(), `unknown parameter "patern"`) {
		t.Fatalf("expected unknown parameter error, got %v", err)
	}
}

func TestDefinition_Build_RejectsWrongParamType(t *testing.T) {
	d, _ := rule.Builtins().Lookup("required-files")
	if _, err := d.Build(map[string]any{"files": "README.md"}, ""); err == nil {
		t.Fatal("expected type error for non-list files param, got nil")
	}
	if _, err := d.Build(map[string]any{"files": []any{"README.md", 3}}, ""); err == nil {
		t.Fatal("expected type error for non-string list item, got nil")
	}
}

func TestDefinition_Build_AppliesConfiguredSeverity(t *testing.T) {
	d, _ := rule.Builtins().Lookup("no-secrets")
	r, err := d.Build(nil, policy.SeverityWarning)
	if err != nil {
		t.Fatal(err)
	}
	if r.Severity() != policy.SeverityWarning {
		t.Errorf("severity = %s, want WARNING", r.Severity())
	}
}

func TestRegistry_RejectsDuplicates(t *testing.T) {
	reg := rule.NewRegistry()
	def := rule.Definition{ID: "x", Factory: func(rule.Params, policy.Severity) (policy.Rule, error) { return nil, nil }}
	if err := reg.Register(def); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(def); err == nil {
		t.Fatal("expected duplicate registration error, got nil")
	}
}

func TestDefinition_Build_ExcludeSkipsFilesAndIssues(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{
		"main.go":              "package main\n",
		"testdata/key.pem":     "-----BEGIN " + "RSA PRIVATE KEY-----\n",
		"server_test.go":       "const k = \"AKIA" + "IOSFODNN7EXAMPLE\"\n",
		"internal/real/env.go": "const k = \"AKIA" + "IOSFODNN7EXAMPLE\"\n",
	} {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	d, _ := rule.Builtins().Lookup("no-secrets")
	r, err := d.Build(map[string]any{"exclude": []any{"**/testdata/**", "*_test.go"}}, "")
	if err != nil {
		t.Fatal(err)
	}
	if r.ID() != "no-secrets" {
		t.Errorf("wrapped rule ID = %q, want no-secrets", r.ID())
	}

	files := []string{"internal/real/env.go", "main.go", "server_test.go", "testdata/key.pem"}
	issues, err := r.Run(policy.NewScanContext(context.Background(), dir, files))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].FilePath != "internal/real/env.go" {
		t.Errorf("expected only internal/real/env.go to be reported, got %+v", issues)
	}
}

func TestDefinition_Build_ExcludeDropsExistenceIssues(t *testing.T) {
	d, _ := rule.Builtins().Lookup("required-files")
	r, err := d.Build(map[string]any{"files": []any{"README.md", "docs/SPEC.md"}, "exclude": []any{"docs"}}, "")
	if err != nil {
		t.Fatal(err)
	}
	issues, err := r.Run(policy.NewScanContext(context.Background(), t.TempDir(), nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].FilePath != "README.md" {
		t.Errorf("expected only README.md issue, got %+v", issues)
	}
}

func TestDefinition_Build_RejectsInvalidExclude(t *testing.T) {
	d, _ := rule.Builtins().Lookup("file-naming")
	if _, err := d.Build(map[string]any{"exclude": "testdata"}, ""); err == nil {
		t.Error("expected error for non-list exclude")
	}
	if _, err := d.Build(map[string]any{"exclude": []any{"bad/[pattern"}}, ""); err == nil {
		t.Error("expected error for malformed exclude pattern")
	}
}
