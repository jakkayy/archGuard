package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var forceInit bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default archguard.yaml configuration file",
	Long:  `Creates a sample archguard.yaml policy configuration file in the current workspace directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := "archguard.yaml"

		if _, err := os.Stat(filename); err == nil && !forceInit {
			yellow := color.New(color.FgYellow).SprintFunc()
			fmt.Printf("%s Config file '%s' already exists. Use '--force' to overwrite.\n", yellow("⚠️"), filename)
			return nil
		}

		defaultContent := `version: "v1"
rules:
  file-naming:
    enabled: true
    severity: WARNING
    pattern: "^[a-z0-9_\\-\\.]+$"

  openapi-exists:
    enabled: true
    severity: ERROR
    path: "docs/openapi.json"
`
		if err := os.WriteFile(filename, []byte(defaultContent), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}

		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		fmt.Printf("%s Created default configuration file '%s' successfully!\n", green("✨"), filename)
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "Overwrite existing archguard.yaml configuration file")
	rootCmd.AddCommand(initCmd)
}
