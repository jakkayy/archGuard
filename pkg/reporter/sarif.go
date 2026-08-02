package reporter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/archguard/archguard/internal/core"
)

// SARIFLog represents the top-level SARIF v2.1.0 JSON structure.
type SARIFLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []SARIFRun `json:"runs"`
}

// SARIFRun represents an execution run of the ArchGuard analysis tool.
type SARIFRun struct {
	Tool    SARIFTool     `json:"tool"`
	Results []SARIFResult `json:"results"`
}

// SARIFTool defines the scanner engine metadata.
type SARIFTool struct {
	Driver SARIFDriver `json:"driver"`
}

// SARIFDriver represents tool driver info.
type SARIFDriver struct {
	Name           string `json:"name"`
	InformationURI string `json:"informationUri"`
}

// SARIFResult represents an individual issue result in SARIF format.
type SARIFResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   SARIFText       `json:"message"`
	Locations []SARIFLocation `json:"locations,omitempty"`
}

// SARIFText holds text content for SARIF messages.
type SARIFText struct {
	Text string `json:"text"`
}

// SARIFLocation holds physical location information of the issue.
type SARIFLocation struct {
	PhysicalLocation SARIFPhysicalLocation `json:"physicalLocation"`
}

// SARIFPhysicalLocation specifies the file artifact path.
type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
}

// SARIFArtifactLocation specifies the relative URI to the file.
type SARIFArtifactLocation struct {
	URI string `json:"uri"`
}

// SARIFReporter formats ScanResult into GitHub-compatible SARIF v2.1.0 JSON.
type SARIFReporter struct{}

// NewSARIFReporter initializes a new SARIFReporter instance.
func NewSARIFReporter() *SARIFReporter {
	return &SARIFReporter{}
}

// Report serializes ScanResult into SARIF v2.1.0 JSON and writes to the provided Writer.
func (r *SARIFReporter) Report(w io.Writer, res *core.ScanResult) error {
	results := make([]SARIFResult, 0)

	for _, issue := range res.Issues {
		level := "note"
		switch issue.Severity {
		case core.SeverityError:
			level = "error"
		case core.SeverityWarning:
			level = "warning"
		}

		msgText := issue.Message
		if issue.Suggestion != "" {
			msgText = fmt.Sprintf("%s. Suggestion: %s", issue.Message, issue.Suggestion)
		}

		sarifRes := SARIFResult{
			RuleID:  issue.RuleID,
			Level:   level,
			Message: SARIFText{Text: msgText},
		}

		if issue.FilePath != "" {
			sarifRes.Locations = []SARIFLocation{
				{
					PhysicalLocation: SARIFPhysicalLocation{
						ArtifactLocation: SARIFArtifactLocation{
							URI: issue.FilePath,
						},
					},
				},
			}
		}

		results = append(results, sarifRes)
	}

	sarifLog := SARIFLog{
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Version: "2.1.0",
		Runs: []SARIFRun{
			{
				Tool: SARIFTool{
					Driver: SARIFDriver{
						Name:           "ArchGuard",
						InformationURI: "https://github.com/archguard/archguard",
					},
				},
				Results: results,
			},
		},
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(sarifLog); err != nil {
		return fmt.Errorf("failed to encode SARIF report: %w", err)
	}

	return nil
}
