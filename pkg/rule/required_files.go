package rule

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/archguard/archguard/internal/core"
)

// RequiredFilesRule validates that designated mandatory files exist in the project.
type RequiredFilesRule struct {
	id            string
	name          string
	description   string
	severity      core.Severity
	requiredFiles []string
}

// NewRequiredFilesRule initializes a RequiredFilesRule with target required files and severity.
func NewRequiredFilesRule(files []string, severity core.Severity) *RequiredFilesRule {
	if len(files) == 0 {
		files = []string{"README.md"}
	}
	if severity == "" {
		severity = core.SeverityError
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
func (r *RequiredFilesRule) Severity() core.Severity {
	return r.severity
}

// Run executes the required files check against the workspace.
func (r *RequiredFilesRule) Run(ctx *core.ScanContext) ([]core.Issue, error) {
	var issues []core.Issue

	for _, relPath := range r.requiredFiles {
		fullPath := filepath.Join(ctx.WorkingDir, relPath)
		info, err := os.Stat(fullPath)
		if os.IsNotExist(err) || (err == nil && info.IsDir()) {
			issues = append(issues, core.Issue{
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
