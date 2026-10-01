package core_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jakkayy/archGuard/internal/config"
	"github.com/jakkayy/archGuard/internal/core"
)

type mockRule struct {
	id       string
	severity core.Severity
	issues   []core.Issue
}

func (m *mockRule) ID() string              { return m.id }
func (m *mockRule) Name() string            { return "Mock Rule" }
func (m *mockRule) Description() string     { return "Mock Rule Description" }
func (m *mockRule) Severity() core.Severity { return m.severity }
func (m *mockRule) Run(ctx *core.ScanContext) ([]core.Issue, error) {
	return m.issues, nil
}

func TestEngine_Run(t *testing.T) {
	eng := core.NewEngine()

	mock := &mockRule{
		id:       "file-naming",
		severity: core.SeverityError,
		issues: []core.Issue{
			{
				RuleID:   "file-naming",
				FilePath: "BadFile.go",
				Message:  "invalid name",
				Severity: core.SeverityError,
			},
		},
	}

	eng.RegisterRule(mock)

	cfg := &config.Config{
		Version: "v1",
		Rules: map[string]config.RuleConfig{
			"file-naming": {
				Enabled:  true,
				Severity: "ERROR",
			},
		},
	}

	res, err := eng.Run(context.Background(), ".", cfg)
	if err != nil {
		t.Fatalf("unexpected error running engine: %v", err)
	}

	if res.Passed {
		t.Error("expected scan to fail due to error severity issue, got passed")
	}

	if len(res.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(res.Issues))
	}

	if res.ErrorCount() != 1 {
		t.Errorf("expected ErrorCount 1, got %d", res.ErrorCount())
	}
}

func TestEngine_Run_UnknownRuleSuggestsClosestMatch(t *testing.T) {
	eng := core.NewEngine()
	eng.RegisterRule(&mockRule{id: "no-secrets", severity: core.SeverityError})

	cfg := &config.Config{
		Rules: map[string]config.RuleConfig{
			"no-secret": {Enabled: true},
		},
	}

	_, err := eng.Run(context.Background(), t.TempDir(), cfg)
	if err == nil {
		t.Fatal("expected error for unknown rule, got nil")
	}
	if !strings.Contains(err.Error(), `did you mean "no-secrets"`) {
		t.Errorf("expected suggestion in error, got: %v", err)
	}
}

func TestEngine_Run_UnknownDisabledRuleStillErrors(t *testing.T) {
	eng := core.NewEngine()
	cfg := &config.Config{
		Rules: map[string]config.RuleConfig{
			"does-not-exist": {Enabled: false},
		},
	}
	if _, err := eng.Run(context.Background(), t.TempDir(), cfg); err == nil {
		t.Fatal("expected error for unknown rule even when disabled, got nil")
	}
}

func TestEngine_Run_IssuesAreSortedDeterministically(t *testing.T) {
	eng := core.NewEngine()
	eng.RegisterRule(&mockRule{id: "rule-b", severity: core.SeverityWarning, issues: []core.Issue{
		{RuleID: "rule-b", FilePath: "b.go", Line: 2, Severity: core.SeverityWarning},
		{RuleID: "rule-b", FilePath: "a.go", Line: 9, Severity: core.SeverityWarning},
	}})
	eng.RegisterRule(&mockRule{id: "rule-a", severity: core.SeverityWarning, issues: []core.Issue{
		{RuleID: "rule-a", FilePath: "b.go", Line: 2, Severity: core.SeverityWarning},
		{RuleID: "rule-a", FilePath: "a.go", Line: 1, Severity: core.SeverityWarning},
	}})

	cfg := &config.Config{Rules: map[string]config.RuleConfig{
		"rule-a": {Enabled: true},
		"rule-b": {Enabled: true},
	}}

	want := []string{"a.go:1:rule-a", "a.go:9:rule-b", "b.go:2:rule-a", "b.go:2:rule-b"}
	for i := 0; i < 20; i++ {
		res, err := eng.Run(context.Background(), t.TempDir(), cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for j, issue := range res.Issues {
			got := fmt.Sprintf("%s:%d:%s", issue.FilePath, issue.Line, issue.RuleID)
			if got != want[j] {
				t.Fatalf("run %d: issue %d = %s, want %s", i, j, got, want[j])
			}
		}
	}
}

func TestEngine_Run_StopsWhenContextCancelled(t *testing.T) {
	eng := core.NewEngine()
	eng.RegisterRule(&mockRule{id: "rule-a", severity: core.SeverityError})
	cfg := &config.Config{Rules: map[string]config.RuleConfig{"rule-a": {Enabled: true}}}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := eng.Run(ctx, t.TempDir(), cfg); err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}
