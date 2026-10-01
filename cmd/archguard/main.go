package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
)

// Exit codes: 0 = scan passed, 1 = ERROR-level policy violations found, 2 = usage/config/runtime error.
const (
	exitOK           = 0
	exitPolicyFailed = 1
	exitError        = 2
)

// errPolicyFailed is returned by scan when ERROR-level violations were found. The report
// has already been printed, so main exits with exitPolicyFailed without an extra message.
var errPolicyFailed = errors.New("policy violations found")

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
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the CLI with args and returns the process exit code.
func run(args []string, stdout, stderr io.Writer) int {
	rootCmd.SetArgs(args)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	if err := rootCmd.Execute(); err != nil {
		if errors.Is(err, errPolicyFailed) {
			return exitPolicyFailed
		}
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return exitError
	}
	return exitOK
}
