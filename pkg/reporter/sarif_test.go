package reporter

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/jakkayy/archGuard/pkg/policy"
)

func TestSARIFReporter_Report(t *testing.T) {
	rep := NewSARIFReporter(nil, "")
	var buf bytes.Buffer

	res := &policy.ScanResult{
		Issues: []policy.Issue{
			{
				RuleID:     "file-naming",
				FilePath:   "src/BadFile.js",
				Message:    "filename does not match pattern",
				Severity:   policy.SeverityWarning,
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

type stubRule struct{ id string }

func (s stubRule) ID() string                                      { return s.id }
func (s stubRule) Name() string                                    { return "Stub " + s.id }
func (s stubRule) Description() string                             { return "stub description" }
func (s stubRule) Severity() policy.Severity                       { return policy.SeverityError }
func (s stubRule) Run(*policy.ScanContext) ([]policy.Issue, error) { return nil, nil }

func TestSARIFReporter_IncludesRegionAndRuleMetadata(t *testing.T) {
	rep := NewSARIFReporter([]policy.Rule{stubRule{id: "no-secrets"}}, "v1.2.3")
	var buf bytes.Buffer

	res := &policy.ScanResult{Issues: []policy.Issue{
		{RuleID: "no-secrets", FilePath: "src/config.go", Line: 42, Message: "secret", Severity: policy.SeverityError},
		{RuleID: "required-files", FilePath: "README.md", Message: "missing", Severity: policy.SeverityError},
	}}
	if err := rep.Report(&buf, res); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var sarif SARIFLog
	if err := json.Unmarshal(buf.Bytes(), &sarif); err != nil {
		t.Fatalf("failed unmarshaling SARIF JSON: %v", err)
	}
	run := sarif.Runs[0]

	if run.Tool.Driver.Version != "v1.2.3" {
		t.Errorf("expected driver version v1.2.3, got %q", run.Tool.Driver.Version)
	}
	if len(run.Tool.Driver.Rules) != 1 || run.Tool.Driver.Rules[0].ID != "no-secrets" {
		t.Fatalf("expected driver rules to contain no-secrets, got %+v", run.Tool.Driver.Rules)
	}

	withLine := run.Results[0]
	if withLine.RuleIndex == nil || *withLine.RuleIndex != 0 {
		t.Errorf("expected ruleIndex 0, got %v", withLine.RuleIndex)
	}
	region := withLine.Locations[0].PhysicalLocation.Region
	if region == nil || region.StartLine != 42 {
		t.Errorf("expected region.startLine 42, got %+v", region)
	}

	noLine := run.Results[1]
	if noLine.Locations[0].PhysicalLocation.Region != nil {
		t.Errorf("expected no region for file-level issue, got %+v", noLine.Locations[0].PhysicalLocation.Region)
	}
	if noLine.RuleIndex != nil {
		t.Errorf("expected no ruleIndex for unknown rule, got %d", *noLine.RuleIndex)
	}
}

func TestSARIFReporter_EmptyResultsSerializeAsArrays(t *testing.T) {
	var buf bytes.Buffer
	if err := NewSARIFReporter(nil, "").Report(&buf, &policy.ScanResult{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !bytes.Contains([]byte(out), []byte(`"results": []`)) || !bytes.Contains([]byte(out), []byte(`"rules": []`)) {
		t.Errorf("expected empty results and rules arrays (not null), got:\n%s", out)
	}
}
