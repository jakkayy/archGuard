package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/archguard/archguard/internal/core"
	"github.com/archguard/archguard/pkg/rule"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "List all available policy rules in ArchGuard",
	Long:  `Displays a detailed catalog of all built-in engineering policy rules supported by ArchGuard.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		engine := core.NewEngine()

		fileNaming, err := rule.NewFileNamingRule("", core.SeverityWarning)
		if err == nil {
			engine.RegisterRule(fileNaming)
		}
		engine.RegisterRule(rule.NewOpenAPIExistsRule("", core.SeverityError))

		rules := engine.Rules()
		sort.Slice(rules, func(i, j int) bool {
			return rules[i].ID() < rules[j].ID()
		})

		boldCyan := color.New(color.Bold, color.FgCyan).SprintFunc()
		boldYellow := color.New(color.Bold, color.FgYellow).SprintFunc()
		boldRed := color.New(color.Bold, color.FgRed).SprintFunc()
		bold := color.New(color.Bold).SprintFunc()

		fmt.Printf("\n🛡️  %s\n", boldCyan("ArchGuard Available Policy Rules"))
		fmt.Println(strings.Repeat("─", 65))

		for i, r := range rules {
			var sevBadge string
			switch r.Severity() {
			case core.SeverityError:
				sevBadge = boldRed("ERROR")
			case core.SeverityWarning:
				sevBadge = boldYellow("WARN")
			default:
				sevBadge = bold("INFO")
			}

			fmt.Printf("%d. [%s] %s\n", i+1, boldCyan(r.ID()), bold(r.Name()))
			fmt.Printf("   • Description: %s\n", r.Description())
			fmt.Printf("   • Default Severity: %s\n", sevBadge)

			switch r.ID() {
			case "file-naming":
				fmt.Printf("   • Parameters:       %s (string, regex pattern)\n", bold("pattern"))
			case "openapi-exists":
				fmt.Printf("   • Parameters:       %s (string, target file path)\n", bold("path"))
			}
			fmt.Println()
		}

		fmt.Printf("Total Available Rules: %d\n\n", len(rules))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rulesCmd)
}
