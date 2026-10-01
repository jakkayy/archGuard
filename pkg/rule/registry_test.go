package rule_test

import (
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
