package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// hookMarker identifies pre-commit hooks written by ArchGuard so they can be safely replaced or removed.
const hookMarker = "# ArchGuard Git Pre-Commit Hook"

const hookScript = `#!/bin/sh
` + hookMarker + `

if ! command -v archguard >/dev/null 2>&1; then
  echo "⚠️  archguard not found in PATH; skipping policy check."
  exit 0
fi

echo "🛡️  Running ArchGuard Policy Check before commit..."
if ! archguard scan; then
  echo "🚨 ArchGuard scan failed! Fix error-level policy violations before committing."
  exit 1
fi
`

var forceHook bool

var installHookCmd = &cobra.Command{
	Use:   "install-hook",
	Short: "Install ArchGuard as a Git pre-commit hook",
	Long:  `Installs a pre-commit Git hook that runs 'archguard scan' before every commit. Works from any directory inside the repository, including worktrees.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		hookPath, err := preCommitHookPath()
		if err != nil {
			return err
		}

		if existing, err := os.ReadFile(hookPath); err == nil && !bytes.Contains(existing, []byte(hookMarker)) && !forceHook {
			return fmt.Errorf("a pre-commit hook not managed by ArchGuard already exists at %s; use --force to overwrite it", hookPath)
		}

		if err := os.MkdirAll(filepath.Dir(hookPath), 0755); err != nil {
			return fmt.Errorf("failed to create git hooks directory %s: %w", filepath.Dir(hookPath), err)
		}
		if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
			return fmt.Errorf("failed to write git pre-commit hook at %s: %w", hookPath, err)
		}
		// WriteFile keeps the mode of an existing file, so make sure it is executable.
		if err := os.Chmod(hookPath, 0755); err != nil {
			return fmt.Errorf("failed to make hook executable: %w", err)
		}

		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		fmt.Fprintf(cmd.OutOrStdout(), "%s Successfully installed ArchGuard pre-commit hook at '%s'!\n", green("✨"), hookPath)
		return nil
	},
}

var uninstallHookCmd = &cobra.Command{
	Use:   "uninstall-hook",
	Short: "Remove the ArchGuard Git pre-commit hook",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		hookPath, err := preCommitHookPath()
		if err != nil {
			return err
		}

		existing, err := os.ReadFile(hookPath)
		if os.IsNotExist(err) {
			fmt.Fprintln(cmd.OutOrStdout(), "No pre-commit hook installed; nothing to do.")
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", hookPath, err)
		}
		if !bytes.Contains(existing, []byte(hookMarker)) {
			return fmt.Errorf("pre-commit hook at %s was not installed by ArchGuard; refusing to remove it", hookPath)
		}

		if err := os.Remove(hookPath); err != nil {
			return fmt.Errorf("failed to remove %s: %w", hookPath, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Removed ArchGuard pre-commit hook at '%s'.\n", hookPath)
		return nil
	},
}

// preCommitHookPath asks git where hooks live, which handles subdirectories, worktrees,
// submodules and a custom core.hooksPath.
func preCommitHookPath() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--git-path", "hooks/pre-commit").Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository (or git is not installed); run 'git init' first")
	}
	p := strings.TrimSpace(string(out))
	if !filepath.IsAbs(p) {
		if p, err = filepath.Abs(p); err != nil {
			return "", err
		}
	}
	return p, nil
}

func init() {
	installHookCmd.Flags().BoolVar(&forceHook, "force", false, "Overwrite an existing pre-commit hook not created by ArchGuard")
	rootCmd.AddCommand(installHookCmd)
	rootCmd.AddCommand(uninstallHookCmd)
}
