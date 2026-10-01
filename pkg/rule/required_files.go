package rule

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jakkayy/archGuard/pkg/policy"
)

// RequiredFilesRule validates that designated mandatory files exist in the project.
type RequiredFilesRule struct {
	id            string
	name          string
	description   string
	severity      policy.Severity
	requiredFiles []string
}

// NewRequiredFilesRule initializes a RequiredFilesRule with target required files and severity.
func NewRequiredFilesRule(files []string, severity policy.Severity) *RequiredFilesRule {
	if len(files) == 0 {
		files = []string{"README.md"}
	}
	if severity == "" {
		severity = policy.SeverityError
	}

	return &RequiredFilesRule{
		id:            "required-files",
		name:          "Required Files Existence Check",
		description:   "Validates that mandatory files exist in the project workspace",
		severity:      severity,
		requiredFiles: files,
	}
}

// ID returns the unique identifier for the rule.
func (r *RequiredFilesRule) ID() string {
	return r.id
}

// Name returns the human-readable name of the rule.
func (r *RequiredFilesRule) Name() string {
	return r.name
}

// Description returns the description of what the rule validates.
func (r *RequiredFilesRule) Description() string {
	return r.description
}

// Severity returns the severity level of rule violations.
func (r *RequiredFilesRule) Severity() policy.Severity {
	return r.severity
}

// Run executes the required files check against the workspace.
func (r *RequiredFilesRule) Run(ctx *policy.ScanContext) ([]policy.Issue, error) {
	var issues []policy.Issue

	for _, relPath := range r.requiredFiles {
		fullPath := filepath.Join(ctx.WorkingDir, relPath)
		info, err := os.Stat(fullPath)
		if os.IsNotExist(err) || (err == nil && info.IsDir()) {
			issues = append(issues, policy.Issue{
				RuleID:     r.id,
				FilePath:   relPath,
				Message:    fmt.Sprintf("required file '%s' was not found in the workspace", relPath),
				Severity:   r.severity,
				Suggestion: fmt.Sprintf("Create mandatory file '%s'", relPath),
			})
		}
	}

	return issues, nil
}
