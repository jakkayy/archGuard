package reporter_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/archguard/archguard/internal/core"
	"github.com/archguard/archguard/pkg/reporter"
)

func TestJSONReporter_Report(t *testing.T) {
	rep := reporter.NewJSONReporter()

	res := &core.ScanResult{
		ScanTimeMs: 20,
		Passed:     true,
		Issues:     []core.Issue{},
	}

	var buf bytes.Buffer
	err := rep.Report(&buf, res)
	if err != nil {
		t.Fatalf("unexpected error from json reporter: %v", err)
	}

	var decoded core.ScanResult
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
