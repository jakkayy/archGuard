package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "archguard",
	Short:   "ArchGuard - Open-source Engineering Policy Engine",
	Long:    `ArchGuard is an automated engineering policy engine that enforces code standards, API compatibility, and architecture boundaries.`,
	Version: version,
}
