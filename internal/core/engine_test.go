package core_test

import (
	"context"
	"testing"

	"github.com/archguard/archguard/internal/config"
	"github.com/archguard/archguard/internal/core"
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
