package core_test

import (
	"context"
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
