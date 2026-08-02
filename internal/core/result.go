package core

// Severity represents the severity level of an architectural or policy violation.
type Severity string

const (
	SeverityError   Severity = "ERROR"
	SeverityWarning Severity = "WARNING"
	SeverityInfo    Severity = "INFO"
)

// Issue represents a single rule violation detected during project scan.
type Issue struct {
	RuleID     string   `json:"rule_id"`
	FilePath   string   `json:"file_path"`
	Line       int      `json:"line,omitempty"`
	Message    string   `json:"message"`
	Severity   Severity `json:"severity"`
	Suggestion string   `json:"suggestion,omitempty"`
}

// ScanResult holds the aggregated output of a complete scan execution.
type ScanResult struct {
	Issues     []Issue `json:"issues"`
	ScanTimeMs int64   `json:"scan_time_ms"`
	Passed     bool    `json:"passed"`
}

// ErrorCount returns the number of issues with ERROR severity.
func (r *ScanResult) ErrorCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == SeverityError {
			count++
		}
	}
	return count
}

// WarningCount returns the number of issues with WARNING severity.
func (r *ScanResult) WarningCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == SeverityWarning {
			count++
		}
	}
	return count
}
