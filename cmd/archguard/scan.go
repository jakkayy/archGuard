package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jakkayy/archGuard/internal/config"
	"github.com/jakkayy/archGuard/internal/core"
	"github.com/jakkayy/archGuard/pkg/policy"
	"github.com/jakkayy/archGuard/pkg/reporter"
	"github.com/jakkayy/archGuard/pkg/rule"
)

var (
	configPath string
	formatFlag string
	noColor    bool
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan project for engineering policy violations",
	Long:  `Scans project files against rules defined in archguard.yaml and outputs a compliance report.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !isValidFormat(formatFlag) {
			return fmt.Errorf("unsupported --format %q (allowed: %s)", formatFlag, strings.Join(supportedFormats, ", "))
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		eng, err := buildEngine(cfg)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		res, err := eng.Run(cmd.Context(), ".", cfg)
		if err != nil {
			return fmt.Errorf("scan execution error: %w", err)
		}

		var rep reporter.Reporter
		switch formatFlag {
		case "json":
			rep = reporter.NewJSONReporter()
		case "sarif":
			rep = reporter.NewSARIFReporter(eng.Rules(), rootCmd.Version)
		default:
			rep = reporter.NewConsoleReporter(noColor)
		}

		if err := rep.Report(cmd.OutOrStdout(), res); err != nil {
			return fmt.Errorf("failed to format report: %w", err)
		}

		if !res.Passed {
			return errPolicyFailed
		}

		return nil
	},
}

var supportedFormats = []string{"console", "json", "sarif"}

func isValidFormat(f string) bool {
	for _, s := range supportedFormats {
		if f == s {
			return true
		}
	}
	return false
}

// buildEngine registers every built-in rule, configured with params from cfg when present.
func buildEngine(cfg *config.Config) (*core.Engine, error) {
	eng := core.NewEngine()
	for _, def := range rule.Builtins().Definitions() {
		rc := cfg.Rules[def.ID]
		r, err := def.Build(rc.Params, policy.Severity(rc.Severity))
		if err != nil {
			return nil, err
		}
		eng.RegisterRule(r)
	}
	return eng, nil
}

func init() {
	scanCmd.Flags().StringVarP(&configPath, "config", "c", "archguard.yaml", "Path to archguard.yaml configuration file")
	scanCmd.Flags().StringVarP(&formatFlag, "format", "f", "console", "Report format (console, json, sarif)")
	scanCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored terminal output")

	rootCmd.AddCommand(scanCmd)
}
