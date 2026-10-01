package rule

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jakkayy/archGuard/pkg/policy"
)

type secretPattern struct {
	name        string
	regex       *regexp.Regexp
	description string
	// valueGroup, when > 0, is the capture group holding the assigned value; matches whose
	// value is clearly not a secret (see isNonSecretValue) are skipped.
	valueGroup int
}

// envVarName matches upper-case identifiers with at least one underscore, e.g.
// TENCENTCLOUD_SECRET_KEY: the value names an environment variable rather than holding a secret.
var envVarName = regexp.MustCompile(`^[A-Z][A-Z0-9]*(_[A-Z0-9]+)+$`)

// isNonSecretValue reports whether an assigned value is an env var name or a
// single repeated character placeholder such as "xxxxxxxxxxxxxxxx".
func isNonSecretValue(v string) bool {
	if envVarName.MatchString(v) {
		return true
	}
	return strings.Count(v, v[:1]) == len(v)
}

// NoSecretsRule scans project files for hardcoded API keys, tokens, and credentials.
type NoSecretsRule struct {
	id          string
	name        string
	description string
	severity    policy.Severity
	patterns    []secretPattern
}

// NewNoSecretsRule initializes a NoSecretsRule with pre-defined security patterns.
func NewNoSecretsRule(severity policy.Severity) *NoSecretsRule {
	if severity == "" {
		severity = policy.SeverityError
	}

	patterns := []secretPattern{
		{
			name:        "AWS Access Key",
			regex:       regexp.MustCompile(`\b(AKIA|ASIA)[0-9A-Z]{16}\b`),
			description: "Potential AWS Access Key ID detected",
		},
		{
			name:        "Private Key",
			regex:       regexp.MustCompile(`-----BEGIN ((RSA|DSA|EC|OPENSSH|PGP|ENCRYPTED) )?PRIVATE KEY( BLOCK)?-----`),
			description: "Private key block detected",
		},
		{
			name:        "GitHub Token",
			regex:       regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9]{36,}\b`),
			description: "Potential GitHub access token detected",
		},
		{
			name:        "Hardcoded API Secret",
			regex:       regexp.MustCompile(`(?i)(api_key|apikey|secret_key|private_key|auth_token|access_token|client_secret)\s*[:=]\s*["']([a-zA-Z0-9_\-]{16,})["']`),
			description: "Potential hardcoded API key or secret token detected",
			valueGroup:  2,
		},
		{
			name:        "Generic Bearer Token",
			regex:       regexp.MustCompile(`(?i)bearer\s+[a-zA-Z0-9_\-\.]{32,}`),
			description: "Hardcoded Bearer authentication token detected",
		},
	}

	return &NoSecretsRule{
		id:          "no-secrets",
		name:        "No Hardcoded Secrets Check",
		description: "Scans project files for hardcoded API keys, private keys, and tokens",
		severity:    severity,
		patterns:    patterns,
	}
}

// ID returns the unique identifier for the rule.
func (r *NoSecretsRule) ID() string {
	return r.id
}

// Name returns the human-readable name of the rule.
func (r *NoSecretsRule) Name() string {
	return r.name
}

// Description returns the description of what the rule validates.
func (r *NoSecretsRule) Description() string {
	return r.description
}

// Severity returns the severity level of rule violations.
func (r *NoSecretsRule) Severity() policy.Severity {
	return r.severity
}

// IgnoreDirective suppresses no-secrets findings on the line where it appears.
const IgnoreDirective = "archguard:ignore"

const maxScanFileSize = 1024 * 1024

// Run scans workspace files line by line for secret patterns.
func (r *NoSecretsRule) Run(ctx *policy.ScanContext) ([]policy.Issue, error) {
	var issues []policy.Issue

	for _, relPath := range ctx.Files {
		if err := ctx.Ctx.Err(); err != nil {
			return nil, err
		}

		ext := strings.ToLower(filepath.Ext(relPath))
		if isIgnoredExt(ext) {
			continue
		}

		fullPath := filepath.Join(ctx.WorkingDir, relPath)
		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() || info.Size() > maxScanFileSize {
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil || isBinary(content) {
			continue
		}

		for i, line := range strings.Split(string(content), "\n") {
			if strings.Contains(line, IgnoreDirective) {
				continue
			}
			for _, p := range r.patterns {
				if p.matches(line) {
					issues = append(issues, policy.Issue{
						RuleID:     r.id,
						FilePath:   relPath,
						Line:       i + 1,
						Message:    p.description,
						Severity:   r.severity,
						Suggestion: "Remove hardcoded secret and use environment variables or a secret manager (or add '" + IgnoreDirective + "' if this is a false positive)",
					})
				}
			}
		}
	}

	return issues, nil
}

func (p secretPattern) matches(line string) bool {
	if p.valueGroup == 0 {
		return p.regex.MatchString(line)
	}
	for _, m := range p.regex.FindAllStringSubmatch(line, -1) {
		if !isNonSecretValue(m[p.valueGroup]) {
			return true
		}
	}
	return false
}

// isBinary reports whether content looks like a binary file (contains a NUL byte in its first 8KB).
func isBinary(content []byte) bool {
	if len(content) > 8000 {
		content = content[:8000]
	}
	return bytes.IndexByte(content, 0) >= 0
}

func isIgnoredExt(ext string) bool {
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".ico", ".pdf", ".zip", ".exe", ".tar", ".gz", ".lock", ".sum":
		return true
	}
	return false
}
