package core

// Rule defines the standard interface contract that all policy inspection rules must implement.
type Rule interface {
	// ID returns the unique identifier of the rule (e.g., "naming/file-convention").
	ID() string

	// Name returns the human-readable name of the rule.
	Name() string

	// Description returns a concise explanation of what the rule validates.
	Description() string

	// Severity returns the default severity level assigned to issues found by this rule.
	Severity() Severity

	// Run executes the rule against the provided scan context and returns detected issues.
	Run(ctx *ScanContext) ([]Issue, error)
}
