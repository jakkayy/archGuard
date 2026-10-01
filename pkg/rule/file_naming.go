package rule

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jakkayy/archGuard/pkg/policy"
)

// DefaultFileNamingPattern is used when no pattern is configured.
const DefaultFileNamingPattern = `^[a-z0-9._-]+$`

// conventionalNames are well-known files whose upper-case names are an ecosystem
// convention (README.md, LICENSE, Makefile, Dockerfile, ...). They are always allowed,
// with or without an extension or suffix (README.md, LICENSE.txt, Dockerfile.dev).
var conventionalNames = []string{
	"README", "LICENSE", "LICENCE", "COPYING", "NOTICE", "AUTHORS", "CONTRIBUTORS",
	"MAINTAINERS", "CODEOWNERS", "OWNERS", "CHANGELOG", "CHANGES", "HISTORY", "RELEASE_NOTES",
	"CONTRIBUTING", "CODE_OF_CONDUCT", "CONDUCT", "SECURITY", "SUPPORT", "GOVERNANCE",
	"FUNDING", "CITATION", "PULL_REQUEST_TEMPLATE", "ISSUE_TEMPLATE",
	"Makefile", "GNUmakefile", "Dockerfile", "Containerfile", "Jenkinsfile", "Vagrantfile",
	"Procfile", "Gemfile", "Rakefile", "Brewfile", "Pipfile", "Justfile", "Tiltfile",
}

func isConventionalName(base string) bool {
	for _, name := range conventionalNames {
		if base == name || strings.HasPrefix(base, name+".") {
			return true
		}
	}
	return false
}

// FileNamingRule validates project filenames against a specified regex pattern.
type FileNamingRule struct {
	id          string
	name        string
	description string
	severity    policy.Severity
	pattern     *regexp.Regexp
}

// NewFileNamingRule initializes a FileNamingRule with a regex pattern and default severity.
func NewFileNamingRule(patternStr string, severity policy.Severity) (*FileNamingRule, error) {
	if patternStr == "" {
		patternStr = DefaultFileNamingPattern
	}
	if severity == "" {
		severity = policy.SeverityWarning
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
func (r *FileNamingRule) Severity() policy.Severity {
	return r.severity
}

// Run executes the file naming validation against the files in ScanContext.
func (r *FileNamingRule) Run(ctx *policy.ScanContext) ([]policy.Issue, error) {
	var issues []policy.Issue

	for _, relPath := range ctx.Files {
		baseName := filepath.Base(relPath)

		// Ignore special hidden files/directories (starting with .)
		if strings.HasPrefix(baseName, ".") || isConventionalName(baseName) {
			continue
		}

		if !r.pattern.MatchString(baseName) {
			issues = append(issues, policy.Issue{
				RuleID:     r.id,
				FilePath:   relPath,
				Message:    fmt.Sprintf("filename '%s' does not match pattern '%s'", baseName, r.pattern.String()),
				Severity:   r.severity,
				Suggestion: fmt.Sprintf("Rename the file so its name matches the configured pattern %s", r.pattern.String()),
			})
		}
	}

	return issues, nil
}
