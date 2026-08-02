package reporter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/archguard/archguard/internal/core"
	"github.com/archguard/archguard/pkg/reporter"
)

func TestConsoleReporter_Report(t *testing.T) {
	rep := reporter.NewConsoleReporter(true)

	res := &core.ScanResult{
		ScanTimeMs: 15,
		Passed:     false,
		Issues: []core.Issue{
			{
				RuleID:     "naming/file-convention",
				FilePath:   "BadName.go",
				Message:    "invalid filename",
				Severity:   core.SeverityError,
				Suggestion: "use snake_case",
			},
		},
	}

	var buf bytes.Buffer
	err := rep.Report(&buf, res)
	if err != nil {
		t.Fatalf("unexpected error from console reporter: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ArchGuard Policy Scan Report") {
		t.Error("expected output to contain title header")
	}

	if !strings.Contains(output, "BadName.go") {
		t.Error("expected output to contain filename")
	}

	if !strings.Contains(output, "Result: FAILED") {
		t.Error("expected output to contain failed status")
	}
}
