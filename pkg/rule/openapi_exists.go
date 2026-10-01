package rule

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jakkayy/archGuard/pkg/policy"
)

// OpenAPIExistsRule validates that the designated OpenAPI specification file exists.
type OpenAPIExistsRule struct {
	id          string
	name        string
	description string
	severity    policy.Severity
	targetPath  string
}

// NewOpenAPIExistsRule initializes an OpenAPIExistsRule with a target spec path and severity.
func NewOpenAPIExistsRule(targetPath string, severity policy.Severity) *OpenAPIExistsRule {
	if targetPath == "" {
		targetPath = "docs/openapi.yaml"
	}
	if severity == "" {
		severity = policy.SeverityError
	}

	return &OpenAPIExistsRule{
		id:          "openapi-exists",
		name:        "OpenAPI Spec Existence Check",
		description: "Validates that the required OpenAPI specification file exists in the workspace",
		severity:    severity,
		targetPath:  targetPath,
	}
}

// ID returns the unique identifier for the rule.
func (r *OpenAPIExistsRule) ID() string {
	return r.id
}

// Name returns the human-readable name of the rule.
func (r *OpenAPIExistsRule) Name() string {
	return r.name
}

// Description returns the description of what the rule validates.
func (r *OpenAPIExistsRule) Description() string {
	return r.description
}

// Severity returns the severity level of rule violations.
func (r *OpenAPIExistsRule) Severity() policy.Severity {
	return r.severity
}

// Run executes the OpenAPI spec file existence check.
func (r *OpenAPIExistsRule) Run(ctx *policy.ScanContext) ([]policy.Issue, error) {
	var issues []policy.Issue

	fullPath := filepath.Join(ctx.WorkingDir, r.targetPath)
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) || (err == nil && info.IsDir()) {
		issues = append(issues, policy.Issue{
			RuleID:     r.id,
			FilePath:   r.targetPath,
			Message:    fmt.Sprintf("required OpenAPI specification file '%s' was not found", r.targetPath),
			Severity:   r.severity,
			Suggestion: fmt.Sprintf("Create an OpenAPI spec file at '%s'", r.targetPath),
		})
	}

	return issues, nil
}
