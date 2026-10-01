package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/jakkayy/archGuard/pkg/policy"
	"github.com/jakkayy/archGuard/pkg/rule"
)

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "List all available policy rules in ArchGuard",
	Long:  `Displays a detailed catalog of all built-in engineering policy rules supported by ArchGuard.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		defs := rule.Builtins().Definitions()

		boldCyan := color.New(color.Bold, color.FgCyan).SprintFunc()
		boldYellow := color.New(color.Bold, color.FgYellow).SprintFunc()
		boldRed := color.New(color.Bold, color.FgRed).SprintFunc()
		bold := color.New(color.Bold).SprintFunc()

		fmt.Fprintf(out, "\n🛡️  %s\n", boldCyan("ArchGuard Available Policy Rules"))
		fmt.Fprintln(out, strings.Repeat("─", 65))

		for i, d := range defs {
			var sevBadge string
			switch d.DefaultSeverity {
			case policy.SeverityError:
				sevBadge = boldRed("ERROR")
			case policy.SeverityWarning:
				sevBadge = boldYellow("WARNING")
			default:
				sevBadge = bold("INFO")
			}

			fmt.Fprintf(out, "%d. [%s] %s\n", i+1, boldCyan(d.ID), bold(d.Name))
			fmt.Fprintf(out, "   • Description:      %s\n", d.Description)
			fmt.Fprintf(out, "   • Default Severity: %s\n", sevBadge)
			if len(d.Params) == 0 {
				fmt.Fprintln(out, "   • Parameters:       (none)")
			}
			for j, p := range d.Params {
				label := "   • Parameters:      "
				if j > 0 {
					label = "                      "
				}
				fmt.Fprintf(out, "%s %s (%s) - %s\n", label, bold(p.Name), p.Type, p.Description)
			}
			fmt.Fprintln(out)
		}

		fmt.Fprintf(out, "Total Available Rules: %d\n\n", len(defs))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rulesCmd)
}
