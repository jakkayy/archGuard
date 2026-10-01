package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jakkayy/archGuard/internal/config"
	"github.com/jakkayy/archGuard/internal/core"
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
	ruleCfg := func(id string) config.RuleConfig { return cfg.Rules[id] }

	fn := ruleCfg("file-naming")
	pattern, err := stringParam(fn, "file-naming", "pattern")
	if err != nil {
		return nil, err
	}
	namingRule, err := rule.NewFileNamingRule(pattern, core.Severity(fn.Severity))
	if err != nil {
		return nil, fmt.Errorf("failed initializing file-naming rule: %w", err)
	}
	eng.RegisterRule(namingRule)

	oa := ruleCfg("openapi-exists")
	path, err := stringParam(oa, "openapi-exists", "path")
	if err != nil {
		return nil, err
	}
	eng.RegisterRule(rule.NewOpenAPIExistsRule(path, core.Severity(oa.Severity)))

	rf := ruleCfg("required-files")
	files, err := stringSliceParam(rf, "required-files", "files")
	if err != nil {
		return nil, err
	}
	eng.RegisterRule(rule.NewRequiredFilesRule(files, core.Severity(rf.Severity)))

	ns := ruleCfg("no-secrets")
	eng.RegisterRule(rule.NewNoSecretsRule(core.Severity(ns.Severity)))

	return eng, nil
}

func stringParam(rc config.RuleConfig, ruleID, key string) (string, error) {
	raw, ok := rc.Params[key]
	if !ok || raw == nil {
		return "", nil
	}
	s, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("rule %q: parameter %q must be a string, got %T", ruleID, key, raw)
	}
	return s, nil
}

func stringSliceParam(rc config.RuleConfig, ruleID, key string) ([]string, error) {
	raw, ok := rc.Params[key]
	if !ok || raw == nil {
		return nil, nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("rule %q: parameter %q must be a list of strings, got %T", ruleID, key, raw)
	}
	out := make([]string, 0, len(items))
	for i, item := range items {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("rule %q: parameter %q[%d] must be a string, got %T", ruleID, key, i, item)
		}
		out = append(out, s)
	}
	return out, nil
}

func init() {
	scanCmd.Flags().StringVarP(&configPath, "config", "c", "archguard.yaml", "Path to archguard.yaml configuration file")
	scanCmd.Flags().StringVarP(&formatFlag, "format", "f", "console", "Report format (console, json, sarif)")
	scanCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored terminal output")

	rootCmd.AddCommand(scanCmd)
}
