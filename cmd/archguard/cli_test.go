package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/jakkayy/archGuard/pkg/policy"
	"github.com/jakkayy/archGuard/pkg/reporter"
)

// runCLI executes the CLI in-process inside dir and returns exit code, stdout and stderr.
func runCLI(t *testing.T, dir string, args ...string) (int, string, string) {
	t.Helper()
	resetFlags(rootCmd)
	t.Chdir(dir)

	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// resetFlags restores every flag to its default, since cobra keeps flag state between Execute calls.
func resetFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
	for _, c := range cmd.Commands() {
		resetFlags(c)
	}
}

func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

const strictConfig = `version: v1
rules:
  no-secrets:
    enabled: true
  required-files:
    enabled: true
    files: ["README.md"]
  file-naming:
    enabled: true
    pattern: '^[a-z0-9._-]+$'
`

func TestScan_PassingProject(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"archguard.yaml": strictConfig,
		"README.md":      "# demo",
		"main.go":        "package main\n",
	})

	code, out, errOut := runCLI(t, dir, "scan", "--no-color")
	if code != exitOK {
		t.Fatalf("exit = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitOK, out, errOut)
	}
	if !strings.Contains(out, "PASSED") {
		t.Errorf("expected PASSED in output, got:\n%s", out)
	}
}

func TestScan_ViolationsReturnExitOneWithJSONReport(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"archguard.yaml": strictConfig,
		"BadName.go":     "package main\n\nvar key = \"AKIA" + "IOSFODNN7EXAMPLE\"\n",
	})

	code, out, _ := runCLI(t, dir, "scan", "--format=json")
	if code != exitPolicyFailed {
		t.Fatalf("exit = %d, want %d\n%s", code, exitPolicyFailed, out)
	}

	var res policy.ScanResult
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, out)
	}

	got := map[string]int{}
	for _, issue := range res.Issues {
		got[issue.RuleID]++
		if issue.RuleID == "no-secrets" && issue.Line != 3 {
			t.Errorf("no-secrets line = %d, want 3", issue.Line)
		}
	}
	for _, id := range []string{"no-secrets", "required-files", "file-naming"} {
		if got[id] != 1 {
			t.Errorf("expected exactly 1 %s issue, got %d (issues: %+v)", id, got[id], res.Issues)
		}
	}
}

func TestScan_SARIFOutputIsValid(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"archguard.yaml": strictConfig,
		"README.md":      "# demo",
		"leak.txt":       "token: Bearer " + strings.Repeat("x", 40) + "\n",
	})

	code, out, _ := runCLI(t, dir, "scan", "--format", "sarif")
	if code != exitPolicyFailed {
		t.Fatalf("exit = %d, want %d", code, exitPolicyFailed)
	}
	var log reporter.SARIFLog
	if err := json.Unmarshal([]byte(out), &log); err != nil {
		t.Fatalf("invalid SARIF: %v", err)
	}
	var secret *reporter.SARIFResult
	for i, r := range log.Runs[0].Results {
		if r.RuleID == "no-secrets" {
			secret = &log.Runs[0].Results[i]
		}
	}
	if secret == nil {
		t.Fatalf("expected a no-secrets result, got %+v", log.Runs[0].Results)
	}
	if region := secret.Locations[0].PhysicalLocation.Region; region == nil || region.StartLine != 1 {
		t.Errorf("expected region.startLine 1, got %+v", region)
	}
	if len(log.Runs[0].Tool.Driver.Rules) != 4 {
		t.Errorf("expected 4 rules in driver catalog, got %d", len(log.Runs[0].Tool.Driver.Rules))
	}
}

func TestScan_UsageAndConfigErrorsReturnExitTwo(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		args    []string
		wantErr string
	}{
		{"missing config", "", []string{"scan"}, "failed to read config"},
		{"bad format", strictConfig, []string{"scan", "--format=xml"}, `unsupported --format "xml"`},
		{"typo rule", "rules:\n  no-secret:\n    enabled: true\n", []string{"scan"}, `did you mean "no-secrets"`},
		{"bad severity", "rules:\n  no-secrets:\n    severity: fatal\n", []string{"scan"}, "invalid severity"},
		{"unknown param", "rules:\n  file-naming:\n    patern: x\n", []string{"scan"}, `unknown parameter "patern"`},
		{"unknown flag", strictConfig, []string{"scan", "--bogus"}, "unknown flag"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.config != "" {
				writeFiles(t, dir, map[string]string{"archguard.yaml": tt.config})
			}
			code, _, errOut := runCLI(t, dir, tt.args...)
			if code != exitError {
				t.Fatalf("exit = %d, want %d (stderr: %s)", code, exitError, errOut)
			}
			if !strings.Contains(errOut, tt.wantErr) {
				t.Errorf("stderr = %q, want it to contain %q", errOut, tt.wantErr)
			}
		})
	}
}

func TestInit_NonInteractiveThenScan(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"README.md": "# demo", ".gitignore": "bin/\n"})

	if code, out, errOut := runCLI(t, dir, "init", "-y"); code != exitOK {
		t.Fatalf("init exit = %d\n%s\n%s", code, out, errOut)
	}

	code, out, _ := runCLI(t, dir, "init", "-y")
	if code != exitOK || !strings.Contains(out, "already exists") {
		t.Errorf("second init should warn and not overwrite; exit=%d out=%s", code, out)
	}

	if code, out, errOut := runCLI(t, dir, "scan", "--no-color"); code != exitOK {
		t.Fatalf("scan of freshly initialized project exit = %d\n%s\n%s", code, out, errOut)
	}
}

func TestRules_ListsAllBuiltins(t *testing.T) {
	code, out, _ := runCLI(t, t.TempDir(), "rules")
	if code != exitOK {
		t.Fatalf("exit = %d", code)
	}
	for _, id := range []string{"file-naming", "no-secrets", "openapi-exists", "required-files"} {
		if !strings.Contains(out, id) {
			t.Errorf("rules output missing %s", id)
		}
	}
	if !strings.Contains(out, "Total Available Rules: 4") {
		t.Errorf("expected total of 4 rules, got:\n%s", out)
	}
}

func TestVersionFlag(t *testing.T) {
	code, out, _ := runCLI(t, t.TempDir(), "--version")
	if code != exitOK || !strings.Contains(out, "archguard version") {
		t.Errorf("--version exit=%d out=%q", code, out)
	}
}

func TestHookLifecycle(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	sub := filepath.Join(dir, "nested", "pkg")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")

	// Installing from a subdirectory must target the repository's hook dir.
	if code, _, errOut := runCLI(t, sub, "install-hook"); code != exitOK {
		t.Fatalf("install-hook exit = %d: %s", code, errOut)
	}
	content, err := os.ReadFile(hookPath)
	if err != nil || !strings.Contains(string(content), hookMarker) {
		t.Fatalf("hook not written correctly: %v\n%s", err, content)
	}

	if code, _, errOut := runCLI(t, dir, "uninstall-hook"); code != exitOK {
		t.Fatalf("uninstall-hook exit = %d: %s", code, errOut)
	}
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Fatalf("expected hook to be removed, stat err = %v", err)
	}

	// A foreign hook must not be overwritten or removed without --force.
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\necho custom\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if code, _, _ := runCLI(t, dir, "install-hook"); code != exitError {
		t.Errorf("install-hook over foreign hook exit = %d, want %d", code, exitError)
	}
	if code, _, _ := runCLI(t, dir, "uninstall-hook"); code != exitError {
		t.Errorf("uninstall-hook of foreign hook exit = %d, want %d", code, exitError)
	}
	if code, _, errOut := runCLI(t, dir, "install-hook", "--force"); code != exitOK {
		t.Errorf("install-hook --force exit = %d: %s", code, errOut)
	}
}

func TestInstallHook_OutsideGitRepo(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	code, _, errOut := runCLI(t, dir, "install-hook")
	if code != exitError || !strings.Contains(errOut, "not inside a git repository") {
		t.Errorf("exit = %d stderr = %q", code, errOut)
	}
}
