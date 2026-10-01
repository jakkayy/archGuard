package main

import (
	"fmt"
	"os"
	"runtime/debug"
)

// version is injected at build time via -ldflags "-X main.version=<tag>".
var version = "dev"

func init() {
	rootCmd.Version = resolveVersion()
}

// resolveVersion falls back to the module version embedded by `go install pkg@version`.
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
