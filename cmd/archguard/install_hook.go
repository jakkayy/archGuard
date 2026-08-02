package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var installHookCmd = &cobra.Command{
	Use:   "install-hook",
	Short: "Install ArchGuard as a Git pre-commit hook",
	Long:  `Installs a pre-commit Git hook in .git/hooks/pre-commit to automatically run archguard scan before every commit.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		gitDir := ".git"
		info, err := os.Stat(gitDir)
		if os.IsNotExist(err) || !info.IsDir() {
			return fmt.Errorf("not a git repository (missing .git directory). Please initialize git repository first")
		}

		hooksDir := filepath.Join(gitDir, "hooks")
		if err := os.MkdirAll(hooksDir, 0755); err != nil {
			return fmt.Errorf("failed to create git hooks directory %s: %w", hooksDir, err)
		}

		hookPath := filepath.Join(hooksDir, "pre-commit")

		hookScript := `#!/bin/sh
# ArchGuard Git Pre-Commit Hook

echo "🛡️  Running ArchGuard Policy Check before commit..."
archguard scan

if [ $? -ne 0 ]; then
  echo "🚨 ArchGuard scan failed! Fix error-level policy violations before committing."
  exit 1
fi
`

		if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
			return fmt.Errorf("failed to write git pre-commit hook at %s: %w", hookPath, err)
		}

		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		fmt.Printf("%s Successfully installed ArchGuard pre-commit hook at '%s'!\n", green("✨"), hookPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installHookCmd)
}
