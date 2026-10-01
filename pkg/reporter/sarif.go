package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/jakkayy/archGuard/internal/core"
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

// SARIFDriver represents tool driver info and the catalog of rules it can report.
type SARIFDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version,omitempty"`
	InformationURI string      `json:"informationUri"`
	Rules          []SARIFRule `json:"rules"`
}

// SARIFRule describes a single reporting rule (reportingDescriptor).
type SARIFRule struct {
	ID                   string             `json:"id"`
	Name                 string             `json:"name"`
	ShortDescription     SARIFText          `json:"shortDescription"`
	DefaultConfiguration SARIFRuleConfigLvl `json:"defaultConfiguration"`
}

// SARIFRuleConfigLvl holds the default level of a rule.
type SARIFRuleConfigLvl struct {
	Level string `json:"level"`
}

// SARIFResult represents an individual issue result in SARIF format.
type SARIFResult struct {
	RuleID    string          `json:"ruleId"`
	RuleIndex *int            `json:"ruleIndex,omitempty"`
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

// SARIFPhysicalLocation specifies the file artifact path and optional line region.
type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
	Region           *SARIFRegion          `json:"region,omitempty"`
}

// SARIFArtifactLocation specifies the URI to the file relative to the source root.
type SARIFArtifactLocation struct {
	URI       string `json:"uri"`
	URIBaseID string `json:"uriBaseId,omitempty"`
}

// SARIFRegion specifies the line range of an issue.
type SARIFRegion struct {
	StartLine int `json:"startLine"`
}

// SARIFReporter formats ScanResult into GitHub-compatible SARIF v2.1.0 JSON.
type SARIFReporter struct {
	rules   []core.Rule
	version string
}

// NewSARIFReporter initializes a SARIFReporter. rules populates tool.driver.rules so
// code scanning UIs can show rule names and descriptions; it may be nil.
func NewSARIFReporter(rules []core.Rule, version string) *SARIFReporter {
	return &SARIFReporter{rules: rules, version: version}
}

func sarifLevel(sev core.Severity) string {
	switch sev {
	case core.SeverityError:
		return "error"
	case core.SeverityWarning:
		return "warning"
	default:
		return "note"
	}
}

// Report serializes ScanResult into SARIF v2.1.0 JSON and writes to the provided Writer.
func (r *SARIFReporter) Report(w io.Writer, res *core.ScanResult) error {
	if res == nil {
		return fmt.Errorf("cannot format nil scan result")
	}

	sarifRules := make([]SARIFRule, 0, len(r.rules))
	ruleIndex := make(map[string]int, len(r.rules))
	for i, rule := range r.rules {
		ruleIndex[rule.ID()] = i
		sarifRules = append(sarifRules, SARIFRule{
			ID:                   rule.ID(),
			Name:                 rule.Name(),
			ShortDescription:     SARIFText{Text: rule.Description()},
			DefaultConfiguration: SARIFRuleConfigLvl{Level: sarifLevel(rule.Severity())},
		})
	}

	results := make([]SARIFResult, 0, len(res.Issues))
	for _, issue := range res.Issues {
		msgText := issue.Message
		if issue.Suggestion != "" {
			msgText = fmt.Sprintf("%s. Suggestion: %s", issue.Message, issue.Suggestion)
		}

		sarifRes := SARIFResult{
			RuleID:  issue.RuleID,
			Level:   sarifLevel(issue.Severity),
			Message: SARIFText{Text: msgText},
		}
		if idx, ok := ruleIndex[issue.RuleID]; ok {
			sarifRes.RuleIndex = &idx
		}

		if issue.FilePath != "" {
			loc := SARIFPhysicalLocation{
				ArtifactLocation: SARIFArtifactLocation{
					URI:       filepath.ToSlash(issue.FilePath),
					URIBaseID: "%SRCROOT%",
				},
			}
			if issue.Line > 0 {
				loc.Region = &SARIFRegion{StartLine: issue.Line}
			}
			sarifRes.Locations = []SARIFLocation{{PhysicalLocation: loc}}
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
						Version:        r.version,
						InformationURI: "https://github.com/jakkayy/archGuard",
						Rules:          sarifRules,
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
