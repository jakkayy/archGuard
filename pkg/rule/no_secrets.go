package rule

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/archguard/archguard/internal/core"
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
			regex:       regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
			description: "Potential AWS Access Key ID detected",
		},
		{
			name:        "Private Key",
			regex:       regexp.MustCompile(`-----BEGIN (RSA|OPENSSH|EC|PGP|PRIVATE) KEY-----`),
			description: "Private RSA/SSH Key detected",
		},
		{
			name:        "Hardcoded API Secret",
			regex:       regexp.MustCompile(`(?i)(api_key|apikey|secret_key|private_key|auth_token)\s*[:=]\s*["'][a-zA-Z0-9_\-]{16,}["']`),
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

// Run scans workspace files for secret patterns.
func (r *NoSecretsRule) Run(ctx *core.ScanContext) ([]core.Issue, error) {
	var issues []core.Issue

	for _, relPath := range ctx.Files {
		// Skip non-code / lock / image files for efficiency
		ext := strings.ToLower(filepath.Ext(relPath))
		if isIgnoredExt(ext) {
			continue
		}

		fullPath := filepath.Join(ctx.WorkingDir, relPath)
		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() || info.Size() > 1024*1024 { // Skip files > 1MB
			continue
		}

		contentBytes, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}
		content := string(contentBytes)

		for _, p := range r.patterns {
			if p.regex.MatchString(content) {
				issues = append(issues, core.Issue{
					RuleID:     r.id,
					FilePath:   relPath,
					Message:    fmt.Sprintf("%s found in file '%s'", p.description, relPath),
					Severity:   r.severity,
					Suggestion: "Remove hardcoded secret and use environment variables or secret manager",
				})
			}
		}
	}

	return issues, nil
}

func isIgnoredExt(ext string) bool {
	ignored := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".gif":  true,
		".ico":  true,
		".pdf":  true,
		".zip":  true,
		".exe":  true,
		".tar":  true,
		".gz":   true,
		".lock": true,
		".sum":  true,
	}
	return ignored[ext]
}
