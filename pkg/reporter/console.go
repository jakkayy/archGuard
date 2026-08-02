package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"

	"github.com/archguard/archguard/internal/core"
)

// ConsoleReporter formats scan results into colored terminal output.
type ConsoleReporter struct {
	noColor bool
}

// NewConsoleReporter initializes a new ConsoleReporter instance.
func NewConsoleReporter(noColor bool) *ConsoleReporter {
	return &ConsoleReporter{
		noColor: noColor,
	}
}

// Report writes colored scan result report to the provided io.Writer.
func (c *ConsoleReporter) Report(w io.Writer, res *core.ScanResult) error {
	if res == nil {
		return fmt.Errorf("cannot format nil scan result")
	}

	if c.noColor {
		color.NoColor = true
	}

	red := color.New(color.FgRed, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow, color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	_, _ = fmt.Fprintln(w, bold("\n🛡️  ArchGuard Policy Scan Report"))
	_, _ = fmt.Fprintln(w, strings.Repeat("─", 50))

	if len(res.Issues) == 0 {
		_, _ = fmt.Fprintln(w, green("✨ No policy violations found! Code conforms to engineering standards."))
	} else {
		for _, issue := range res.Issues {
			var badge string
			switch issue.Severity {
			case core.SeverityError:
				badge = red("🚨 ERROR")
			case core.SeverityWarning:
				badge = yellow("⚠️  WARN ")
			default:
				badge = cyan("ℹ️  INFO ")
			}

			loc := issue.FilePath
			if issue.Line > 0 {
				loc = fmt.Sprintf("%s:%d", issue.FilePath, issue.Line)
			}

			_, _ = fmt.Fprintf(w, "[%s] %s (%s)\n", badge, issue.Message, bold(loc))
			_, _ = fmt.Fprintf(w, "       Rule: %s\n", issue.RuleID)
			if issue.Suggestion != "" {
				_, _ = fmt.Fprintf(w, "       Suggestion: %s\n", issue.Suggestion)
			}
			_, _ = fmt.Fprintln(w)
		}
	}

	_, _ = fmt.Fprintln(w, strings.Repeat("─", 50))
	summary := fmt.Sprintf("Scan Time: %d ms | Errors: %d | Warnings: %d", res.ScanTimeMs, res.ErrorCount(), res.WarningCount())
	_, _ = fmt.Fprintln(w, summary)

	if res.Passed {
		_, _ = fmt.Fprintln(w, green("Result: PASSED ✅"))
	} else {
		_, _ = fmt.Fprintln(w, red("Result: FAILED ❌ (Fix ERROR level issues before merging)"))
	}

	return nil
}
