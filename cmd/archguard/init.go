package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	forceInit      bool
	nonInteractive bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default archguard.yaml configuration file",
	Long:  `Creates an archguard.yaml policy configuration file tailored to your project via an interactive setup wizard.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := "archguard.yaml"

		if _, err := os.Stat(filename); err == nil && !forceInit {
			yellow := color.New(color.FgYellow).SprintFunc()
			fmt.Printf("%s Config file '%s' already exists. Use '--force' to overwrite.\n", yellow("⚠️"), filename)
			return nil
		}

		if nonInteractive {
			return writeDefaultConfig(filename)
		}

		bold := color.New(color.Bold, color.FgCyan).SprintFunc()
		fmt.Printf("\n🛡️  %s\n", bold("ArchGuard Interactive Project Setup"))
		fmt.Println(strings.Repeat("─", 50))

		// Question 1: Select Project Category
		catChoices := []string{
			"Frontend / Web App (e.g., Next.js, React, Vue, Svelte)",
			"Backend REST API (e.g., Go, Node.js, Python, Spring Boot)",
			"Full-Stack App (Frontend + Backend in single repository)",
			"Library / CLI Tool (Reusable package or command-line utility)",
		}
		catSelect := selection.New("Select your project category:", catChoices)
		selectedCat, err := catSelect.RunPrompt()
		if err != nil {
			return fmt.Errorf("setup cancelled: %w", err)
		}

		var selectedFramework string
		var isFrontend, isBackend, isFullStack bool

		if strings.HasPrefix(selectedCat, "Frontend") {
			isFrontend = true
			fwChoices := []string{
				"Next.js (App Router / Pages)",
				"React / Vite / Vue / Nuxt / Svelte",
				"Other / Generic HTML & JS",
			}
			fwSelect := selection.New("Select your Frontend Framework:", fwChoices)
			selectedFramework, err = fwSelect.RunPrompt()
			if err != nil {
				return fmt.Errorf("setup cancelled: %w", err)
			}
		} else if strings.HasPrefix(selectedCat, "Backend") {
			isBackend = true
			fwChoices := []string{
				"Go (Gin / Fiber / Echo / Standard)",
				"Node.js (NestJS / Express)",
				"Python (FastAPI / Django / Flask)",
				"Java / Kotlin (Spring Boot)",
			}
			fwSelect := selection.New("Select your Backend Framework:", fwChoices)
			selectedFramework, err = fwSelect.RunPrompt()
			if err != nil {
				return fmt.Errorf("setup cancelled: %w", err)
			}
		} else if strings.HasPrefix(selectedCat, "Full-Stack") {
			isFullStack = true
			isFrontend = true
			isBackend = true

			fwChoicesFrontend := []string{
				"Next.js (App Router / Pages)",
				"React / Vite / Vue / Svelte",
				"Other / Generic HTML",
			}
			fwSelectFE := selection.New("Select your Frontend Framework:", fwChoicesFrontend)
			selectedFE, err := fwSelectFE.RunPrompt()
			if err != nil {
				return fmt.Errorf("setup cancelled: %w", err)
			}

			fwChoicesBackend := []string{
				"Go (Gin / Fiber / Echo / Standard)",
				"Node.js (NestJS / Express)",
				"Python (FastAPI / Django / Flask)",
				"Java / Kotlin (Spring Boot)",
			}
			fwSelectBE := selection.New("Select your Backend Framework:", fwChoicesBackend)
			selectedBE, err := fwSelectBE.RunPrompt()
			if err != nil {
				return fmt.Errorf("setup cancelled: %w", err)
			}

			selectedFramework = fmt.Sprintf("%s + %s", selectedFE, selectedBE)
		} else {
			fwChoices := []string{
				"Go",
				"TypeScript / JavaScript",
				"Python / Other",
			}
			fwSelect := selection.New("Select your primary programming language:", fwChoices)
			selectedFramework, err = fwSelect.RunPrompt()
			if err != nil {
				return fmt.Errorf("setup cancelled: %w", err)
			}
		}

		// Question 3: OpenAPI Specification Check (Only for Backend or Full-Stack)
		openAPIEnabled := false
		openAPIPath := "docs/openapi.json"

		if isBackend {
			openAPIChoices := []string{
				"Yes - Require spec file (Triggers 🚨 ERROR if missing)",
				"No  - Disable OpenAPI check for now",
			}
			openAPISelect := selection.New("Require an OpenAPI / Swagger spec file check for Backend API?", openAPIChoices)
			selectedOpenAPI, err := openAPISelect.RunPrompt()
			if err != nil {
				return fmt.Errorf("setup cancelled: %w", err)
			}

			if strings.HasPrefix(selectedOpenAPI, "Yes") {
				openAPIEnabled = true
				input := textinput.New("Specify OpenAPI spec file path:")
				input.InitialValue = "docs/openapi.json"
				inputPath, err := input.RunPrompt()
				if err == nil && strings.TrimSpace(inputPath) != "" {
					openAPIPath = strings.TrimSpace(inputPath)
				}
			}
		}

		// Question 4: File Naming Policy
		namingChoices := []string{
			"Strict Lowercase (a-z, 0-9, . _ -)      [Recommended for Go / Backend]",
			"Flexible Framework (Include A-Z, [ ] ()) [Recommended for Next.js / React / Full-Stack]",
			"Disabled           (Do not enforce file naming convention)",
		}

		namingSelect := selection.New("Select file naming policy rule:", namingChoices)
		selectedNaming, err := namingSelect.RunPrompt()
		if err != nil {
			return fmt.Errorf("setup cancelled: %w", err)
		}

		namingEnabled := true
		namingPattern := "^[a-z0-9._-]+$"

		if strings.HasPrefix(selectedNaming, "Flexible") || isFrontend || isFullStack {
			if !strings.HasPrefix(selectedNaming, "Disabled") && !strings.HasPrefix(selectedNaming, "Strict") {
				namingPattern = "^[a-zA-Z0-9._\\-\\[\\]\\(\\)]+$"
			}
		}

		if strings.HasPrefix(selectedNaming, "Disabled") {
			namingEnabled = false
		} else if strings.HasPrefix(selectedNaming, "Flexible") {
			namingPattern = "^[a-zA-Z0-9._\\-\\[\\]\\(\\)]+$"
		}

		// Construct archguard.yaml content
		configContent := fmt.Sprintf(`version: "v1"

# Generated for: %s (%s)

rules:
  file-naming:
    enabled: %t
    severity: WARNING
    pattern: "%s"

  openapi-exists:
    enabled: %t
    severity: ERROR
    path: "%s"
`, selectedCat, selectedFramework, namingEnabled, namingPattern, openAPIEnabled, openAPIPath)

		if err := os.WriteFile(filename, []byte(configContent), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}

		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		fmt.Printf("\n%s Created customized configuration file '%s' successfully!\n", green("✨"), filename)
		return nil
	},
}

func writeDefaultConfig(filename string) error {
	defaultContent := `version: "v1"

rules:
  file-naming:
    enabled: true
    severity: WARNING
    pattern: "^[a-zA-Z0-9._-]+$"

  openapi-exists:
    enabled: false
    severity: ERROR
    path: "docs/openapi.json"
`
	if err := os.WriteFile(filename, []byte(defaultContent), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filename, err)
	}

	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	fmt.Printf("%s Created default configuration file '%s' successfully!\n", green("✨"), filename)
	return nil
}

func init() {
	initCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "Overwrite existing archguard.yaml configuration file")
	initCmd.Flags().BoolVarP(&nonInteractive, "non-interactive", "y", false, "Create default configuration without interactive wizard")
	rootCmd.AddCommand(initCmd)
}
