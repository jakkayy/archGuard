package reporter

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/archguard/archguard/internal/core"
)

func TestSARIFReporter_Report(t *testing.T) {
	rep := NewSARIFReporter()
	var buf bytes.Buffer

	res := &core.ScanResult{
		Issues: []core.Issue{
			{
				RuleID:     "file-naming",
				FilePath:   "src/BadFile.js",
				Message:    "filename does not match pattern",
				Severity:   core.SeverityWarning,
				Suggestion: "Rename file",
			},
		},
		ScanTimeMs: 10,
		Passed:     true,
	}

	if err := rep.Report(&buf, res); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var sarif SARIFLog
	if err := json.Unmarshal(buf.Bytes(), &sarif); err != nil {
		t.Fatalf("failed unmarshaling SARIF JSON: %v", err)
	}

	if sarif.Version != "2.1.0" {
		t.Errorf("expected version 2.1.0, got %s", sarif.Version)
	}

	if len(sarif.Runs) != 1 || len(sarif.Runs[0].Results) != 1 {
		t.Fatalf("expected 1 run result, got %d", len(sarif.Runs[0].Results))
	}

	resItem := sarif.Runs[0].Results[0]
	if resItem.RuleID != "file-naming" {
		t.Errorf("expected ruleId file-naming, got %s", resItem.RuleID)
	}
	if resItem.Level != "warning" {
		t.Errorf("expected level warning, got %s", resItem.Level)
	}
}
