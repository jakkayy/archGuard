package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/archguard/archguard/internal/config"
	"github.com/archguard/archguard/internal/core"
	"github.com/archguard/archguard/pkg/reporter"
	"github.com/archguard/archguard/pkg/rule"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		eng := core.NewEngine()

		// Register built-in rules
		if fileNamingCfg, ok := cfg.Rules["file-naming"]; ok {
			pattern, _ := fileNamingCfg.Params["pattern"].(string)
			namingRule, err := rule.NewFileNamingRule(pattern, core.Severity(fileNamingCfg.Severity))
			if err != nil {
				return fmt.Errorf("failed initializing file-naming rule: %w", err)
			}
			eng.RegisterRule(namingRule)
		}

		if openAPICfg, ok := cfg.Rules["openapi-exists"]; ok {
			path, _ := openAPICfg.Params["path"].(string)
			openAPIRule := rule.NewOpenAPIExistsRule(path, core.Severity(openAPICfg.Severity))
			eng.RegisterRule(openAPIRule)
		}

		if reqFilesCfg, ok := cfg.Rules["required-files"]; ok {
			var filesList []string
			if rawFiles, exists := reqFilesCfg.Params["files"].([]any); exists {
				for _, f := range rawFiles {
					if str, isStr := f.(string); isStr {
						filesList = append(filesList, str)
					}
				}
			}
			eng.RegisterRule(rule.NewRequiredFilesRule(filesList, core.Severity(reqFilesCfg.Severity)))
		}

		if noSecretsCfg, ok := cfg.Rules["no-secrets"]; ok {
			eng.RegisterRule(rule.NewNoSecretsRule(core.Severity(noSecretsCfg.Severity)))
		}

		res, err := eng.Run(cmd.Context(), ".", cfg)
		if err != nil {
			return fmt.Errorf("scan execution error: %w", err)
		}

		var rep reporter.Reporter
		switch formatFlag {
		case "json":
			rep = reporter.NewJSONReporter()
		default:
			rep = reporter.NewConsoleReporter(noColor)
		}

		if err := rep.Report(os.Stdout, res); err != nil {
			return fmt.Errorf("failed to format report: %w", err)
		}

		if !res.Passed {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	scanCmd.Flags().StringVarP(&configPath, "config", "c", "archguard.yaml", "Path to archguard.yaml configuration file")
	scanCmd.Flags().StringVarP(&formatFlag, "format", "f", "console", "Report format (console, json)")
	scanCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored terminal output")

	rootCmd.AddCommand(scanCmd)
}
