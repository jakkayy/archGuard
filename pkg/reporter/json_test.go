package reporter_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/jakkayy/archGuard/pkg/policy"
	"github.com/jakkayy/archGuard/pkg/reporter"
)

func TestJSONReporter_Report(t *testing.T) {
	rep := reporter.NewJSONReporter()

	res := &policy.ScanResult{
		ScanTimeMs: 20,
		Passed:     true,
		Issues:     []policy.Issue{},
	}

	var buf bytes.Buffer
	err := rep.Report(&buf, res)
	if err != nil {
		t.Fatalf("unexpected error from json reporter: %v", err)
	}

	var decoded policy.ScanResult
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to parse generated json output: %v", err)
	}

	if decoded.ScanTimeMs != 20 {
		t.Errorf("expected ScanTimeMs 20, got %d", decoded.ScanTimeMs)
	}

	if !decoded.Passed {
		t.Error("expected Passed to be true")
	}
}

func TestJSONReporter_NilIssuesSerializeAsEmptyArray(t *testing.T) {
	var buf bytes.Buffer
	res := &policy.ScanResult{Passed: true}
	if err := reporter.NewJSONReporter().Report(&buf, res); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(`"issues": []`)) {
		t.Errorf("expected \"issues\": [], got:\n%s", buf.String())
	}
	if res.Issues != nil {
		t.Error("reporter must not mutate the caller's ScanResult")
	}
}
