package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/jakkayy/archGuard/internal/config"
)

const (
	strictNamingPattern   = `^[a-z0-9._-]+$`
	flexibleNamingPattern = `^[a-zA-Z0-9._\-\[\]\(\)]+$`
	defaultNamingPattern  = `^[a-zA-Z0-9._-]+$`
	defaultOpenAPIPath    = "docs/openapi.json"
)

var (
	forceInit      bool
	nonInteractive bool
)

// initAnswers captures the wizard choices used to render archguard.yaml.
type initAnswers struct {
	Category       string
	Framework      string
	NamingEnabled  bool
	NamingPattern  string
	OpenAPIEnabled bool
	OpenAPIPath    string
	RequiredFiles  []string
}

func defaultAnswers() initAnswers {
	return initAnswers{
		NamingEnabled: true,
		NamingPattern: defaultNamingPattern,
		OpenAPIPath:   defaultOpenAPIPath,
		RequiredFiles: []string{"README.md", ".gitignore"},
	}
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default archguard.yaml configuration file",
	Long:  `Creates an archguard.yaml policy configuration file tailored to your project via an interactive setup wizard.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := "archguard.yaml"
		out := cmd.OutOrStdout()

		if _, err := os.Stat(filename); err == nil && !forceInit {
			yellow := color.New(color.FgYellow).SprintFunc()
			fmt.Fprintf(out, "%s Config file '%s' already exists. Use '--force' to overwrite.\n", yellow("⚠️"), filename)
			return nil
		}

		answers := defaultAnswers()
		if !nonInteractive {
			bold := color.New(color.Bold, color.FgCyan).SprintFunc()
			fmt.Fprintf(out, "\n🛡️  %s\n", bold("ArchGuard Interactive Project Setup"))
			fmt.Fprintln(out, strings.Repeat("─", 50))

			var err error
			if answers, err = runWizard(); err != nil {
				return fmt.Errorf("setup cancelled: %w (use --non-interactive to skip the wizard)", err)
			}
		}

		content, err := renderConfig(answers)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filename, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}

		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		fmt.Fprintf(out, "%s Created configuration file '%s' successfully!\n", green("✨"), filename)
		return nil
	},
}

func choose(prompt string, choices []string) (string, error) {
	return selection.New(prompt, choices).RunPrompt()
}

var (
	frontendChoices = []string{
		"Next.js (App Router / Pages)",
		"React / Vite / Vue / Nuxt / Svelte",
		"Other / Generic HTML & JS",
	}
	backendChoices = []string{
		"Go (Gin / Fiber / Echo / Standard)",
		"Node.js (NestJS / Express)",
		"Python (FastAPI / Django / Flask)",
		"Java / Kotlin (Spring Boot)",
	}
)

func runWizard() (initAnswers, error) {
	a := defaultAnswers()

	category, err := choose("Select your project category:", []string{
		"Frontend / Web App (e.g., Next.js, React, Vue, Svelte)",
		"Backend REST API (e.g., Go, Node.js, Python, Spring Boot)",
		"Full-Stack App (Frontend + Backend in single repository)",
		"Library / CLI Tool (Reusable package or command-line utility)",
	})
	if err != nil {
		return a, err
	}
	a.Category = category

	isBackend := false
	switch {
	case strings.HasPrefix(category, "Frontend"):
		if a.Framework, err = choose("Select your Frontend Framework:", frontendChoices); err != nil {
			return a, err
		}
	case strings.HasPrefix(category, "Backend"):
		isBackend = true
		if a.Framework, err = choose("Select your Backend Framework:", backendChoices); err != nil {
			return a, err
		}
	case strings.HasPrefix(category, "Full-Stack"):
		isBackend = true
		fe, err := choose("Select your Frontend Framework:", frontendChoices)
		if err != nil {
			return a, err
		}
		be, err := choose("Select your Backend Framework:", backendChoices)
		if err != nil {
			return a, err
		}
		a.Framework = fe + " + " + be
	default:
		if a.Framework, err = choose("Select your primary programming language:", []string{
			"Go",
			"TypeScript / JavaScript",
			"Python / Other",
		}); err != nil {
			return a, err
		}
	}

	if isBackend {
		openAPI, err := choose("Require an OpenAPI / Swagger spec file check for Backend API?", []string{
			"Yes - Require spec file (Triggers 🚨 ERROR if missing)",
			"No  - Disable OpenAPI check for now",
		})
		if err != nil {
			return a, err
		}
		if strings.HasPrefix(openAPI, "Yes") {
			a.OpenAPIEnabled = true
			input := textinput.New("Specify OpenAPI spec file path:")
			input.InitialValue = defaultOpenAPIPath
			p, err := input.RunPrompt()
			if err != nil {
				return a, err
			}
			if p = strings.TrimSpace(p); p != "" {
				a.OpenAPIPath = p
			}
		}
	}

	naming, err := choose("Select file naming policy rule:", []string{
		"Strict Lowercase (a-z, 0-9, . _ -)      [Recommended for Go / Backend]",
		"Flexible Framework (Include A-Z, [ ] ()) [Recommended for Next.js / React / Full-Stack]",
		"Disabled           (Do not enforce file naming convention)",
	})
	if err != nil {
		return a, err
	}
	switch {
	case strings.HasPrefix(naming, "Strict"):
		a.NamingPattern = strictNamingPattern
	case strings.HasPrefix(naming, "Flexible"):
		a.NamingPattern = flexibleNamingPattern
	default:
		a.NamingEnabled = false
	}

	return a, nil
}

// renderConfig marshals the answers into archguard.yaml content. Using the YAML encoder
// (rather than string templates) guarantees user-supplied values are quoted correctly.
func renderConfig(a initAnswers) ([]byte, error) {
	cfg := config.Config{
		Version: config.SupportedVersion,
		Rules: map[string]config.RuleConfig{
			"file-naming": {
				Enabled:  a.NamingEnabled,
				Severity: "WARNING",
				Params:   map[string]any{"pattern": a.NamingPattern},
			},
			"no-secrets": {
				Enabled:  true,
				Severity: "ERROR",
			},
			"required-files": {
				Enabled:  true,
				Severity: "ERROR",
				Params:   map[string]any{"files": a.RequiredFiles},
			},
			"openapi-exists": {
				Enabled:  a.OpenAPIEnabled,
				Severity: "ERROR",
				Params:   map[string]any{"path": a.OpenAPIPath},
			},
		},
	}

	var buf bytes.Buffer
	buf.WriteString("# ArchGuard policy configuration. Run 'archguard rules' to see all available rules.\n")
	if a.Category != "" {
		fmt.Fprintf(&buf, "# Generated for: %s (%s)\n", a.Category, a.Framework)
	}
	buf.WriteString("\n")

	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(cfg); err != nil {
		return nil, fmt.Errorf("failed to render config: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("failed to render config: %w", err)
	}
	return buf.Bytes(), nil
}

func init() {
	initCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "Overwrite existing archguard.yaml configuration file")
	initCmd.Flags().BoolVarP(&nonInteractive, "non-interactive", "y", false, "Create default configuration without interactive wizard")
	rootCmd.AddCommand(initCmd)
}
