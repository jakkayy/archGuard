package rule

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/archguard/archguard/internal/core"
)

// FileNamingRule validates project filenames against a specified regex pattern.
type FileNamingRule struct {
	id          string
	name        string
	description string
	severity    core.Severity
	pattern     *regexp.Regexp
}

// NewFileNamingRule initializes a FileNamingRule with a regex pattern and default severity.
func NewFileNamingRule(patternStr string, severity core.Severity) (*FileNamingRule, error) {
	if patternStr == "" {
		patternStr = `^[a-z0-9._-]+$`
	}
	if severity == "" {
		severity = core.SeverityWarning
	}

	re, err := regexp.Compile(patternStr)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern '%s' for file-naming rule: %w", patternStr, err)
	}

	return &FileNamingRule{
		id:          "file-naming",
		name:        "File Naming Convention",
		description: "Validates that project filenames conform to the specified regex pattern",
		severity:    severity,
		pattern:     re,
	}, nil
}

// ID returns the unique identifier for the rule.
func (r *FileNamingRule) ID() string {
	return r.id
}

// Name returns the human-readable name of the rule.
func (r *FileNamingRule) Name() string {
	return r.name
}

// Description returns the description of what the rule validates.
func (r *FileNamingRule) Description() string {
	return r.description
}

// Severity returns the severity level of rule violations.
func (r *FileNamingRule) Severity() core.Severity {
	return r.severity
}

// Run executes the file naming validation against the files in ScanContext.
func (r *FileNamingRule) Run(ctx *core.ScanContext) ([]core.Issue, error) {
	var issues []core.Issue

	for _, relPath := range ctx.Files {
		baseName := filepath.Base(relPath)

		// Ignore special hidden files/directories (starting with .)
		if strings.HasPrefix(baseName, ".") {
			continue
		}

		if !r.pattern.MatchString(baseName) {
			issues = append(issues, core.Issue{
				RuleID:     r.id,
				FilePath:   relPath,
				Message:    fmt.Sprintf("filename '%s' does not match pattern '%s'", baseName, r.pattern.String()),
				Severity:   r.severity,
				Suggestion: "Rename file using lowercase alphanumeric characters, underscores, or hyphens",
			})
		}
	}

	return issues, nil
}
