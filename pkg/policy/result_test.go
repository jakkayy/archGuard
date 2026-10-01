package policy_test

import (
	"testing"

	"github.com/jakkayy/archGuard/pkg/policy"
)

func TestScanResult_Counts(t *testing.T) {
	res := &policy.ScanResult{Issues: []policy.Issue{
		{Severity: policy.SeverityError},
		{Severity: policy.SeverityWarning},
		{Severity: policy.SeverityWarning},
		{Severity: policy.SeverityInfo},
	}}
	if got := res.ErrorCount(); got != 1 {
		t.Errorf("ErrorCount = %d, want 1", got)
	}
	if got := res.WarningCount(); got != 2 {
		t.Errorf("WarningCount = %d, want 2", got)
	}
}

func TestNewScanContext_DefaultsNilContext(t *testing.T) {
	//nolint:staticcheck // passing nil is the behaviour under test
	sc := policy.NewScanContext(nil, "/repo", []string{"a.go"})
	if sc.Ctx == nil {
		t.Fatal("expected a non-nil background context")
	}
}
