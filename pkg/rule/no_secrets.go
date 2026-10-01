package rule

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jakkayy/archGuard/internal/core"
)

type secretPattern struct {
	name        string
	regex       *regexp.Regexp
	description string
}

// NoSecretsRule scans project files for hardcoded API keys, tokens, and credentials.
type NoSecretsRule struct {
	id          string
	name        string
	description string
	severity    core.Severity
	patterns    []secretPattern
}

// NewNoSecretsRule initializes a NoSecretsRule with pre-defined security patterns.
func NewNoSecretsRule(severity core.Severity) *NoSecretsRule {
	if severity == "" {
		severity = core.SeverityError
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
			regex:       regexp.MustCompile(`(?i)(api_key|apikey|secret_key|private_key|auth_token|access_token|client_secret)\s*[:=]\s*["'][a-zA-Z0-9_\-]{16,}["']`),
			description: "Potential hardcoded API key or secret token detected",
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
func (r *NoSecretsRule) Severity() core.Severity {
	return r.severity
}

// IgnoreDirective suppresses no-secrets findings on the line where it appears.
const IgnoreDirective = "archguard:ignore"

const maxScanFileSize = 1024 * 1024

// Run scans workspace files line by line for secret patterns.
func (r *NoSecretsRule) Run(ctx *core.ScanContext) ([]core.Issue, error) {
	var issues []core.Issue

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
				if p.regex.MatchString(line) {
					issues = append(issues, core.Issue{
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
