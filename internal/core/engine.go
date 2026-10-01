package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jakkayy/archGuard/internal/config"
)

// Engine orchestrates project scanning by executing registered policy rules.
type Engine struct {
	rules map[string]Rule
}

// NewEngine initializes a new Engine instance.
func NewEngine() *Engine {
	return &Engine{
		rules: make(map[string]Rule),
	}
}

// RegisterRule registers a Rule implementation into the engine.
func (e *Engine) RegisterRule(r Rule) {
	if r != nil {
		e.rules[r.ID()] = r
	}
}

// Rules returns a slice of all registered rules in the engine.
func (e *Engine) Rules() []Rule {
	var list []Rule
	for _, r := range e.rules {
		list = append(list, r)
	}
	return list
}

// Run executes all active rules enabled in Config against the target working directory.
func (e *Engine) Run(ctx context.Context, workingDir string, cfg *config.Config) (*ScanResult, error) {
	startTime := time.Now()

	if workingDir == "" {
		workingDir = "."
	}

	absWorkingDir, err := filepath.Abs(workingDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for working directory %s: %w", workingDir, err)
	}

	var customIgnores []string
	if cfg != nil {
		customIgnores = cfg.Ignore
	}

	files, err := e.collectFiles(absWorkingDir, customIgnores)
	if err != nil {
		return nil, fmt.Errorf("failed to collect project files in %s: %w", absWorkingDir, err)
	}

	scanCtx := NewScanContext(ctx, absWorkingDir, files)

	var allIssues []Issue

	if cfg != nil {
		for ruleID, ruleCfg := range cfg.Rules {
			r, ok := e.rules[ruleID]
			if !ok {
				return nil, e.unknownRuleError(ruleID)
			}

			if !ruleCfg.Enabled {
				continue
			}

			issues, err := r.Run(scanCtx)
			if err != nil {
				return nil, fmt.Errorf("failed executing rule %s: %w", ruleID, err)
			}

			allIssues = append(allIssues, issues...)
		}
	}

	scanDuration := time.Since(startTime).Milliseconds()

	result := &ScanResult{
		Issues:     allIssues,
		ScanTimeMs: scanDuration,
		Passed:     true,
	}

	if result.ErrorCount() > 0 {
		result.Passed = false
	}

	return result, nil
}

func (e *Engine) unknownRuleError(ruleID string) error {
	best, bestDist := "", -1
	for id := range e.rules {
		if d := levenshtein(ruleID, id); bestDist < 0 || d < bestDist || (d == bestDist && id < best) {
			best, bestDist = id, d
		}
	}
	if best != "" && bestDist <= 3 {
		return fmt.Errorf("unknown rule %q in config (did you mean %q?)", ruleID, best)
	}
	return fmt.Errorf("unknown rule %q in config", ruleID)
}

func levenshtein(a, b string) int {
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(b)]
}

func (e *Engine) collectFiles(rootDir string, customIgnores []string) ([]string, error) {
	var files []string

	ignoredDirs := map[string]bool{
		".git":          true,
		"node_modules":  true,
		"vendor":        true,
		"bin":           true,
		".next":         true,
		".nuxt":         true,
		".svelte-kit":   true,
		"dist":          true,
		"build":         true,
		"out":           true,
		".output":       true,
		"coverage":      true,
		".cache":        true,
		".turbo":        true,
		"__pycache__":   true,
		".pytest_cache": true,
		".venv":         true,
		"venv":          true,
		"env":           true,
		".mypy_cache":   true,
		"target":        true,
		".gradle":       true,
		".dart_tool":    true,
		".idea":         true,
		".vscode":       true,
	}

	for _, customDir := range customIgnores {
		if customDir != "" {
			ignoredDirs[customDir] = true
		}
	}

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		if d.IsDir() {
			if ignoredDirs[d.Name()] && relPath != "." {
				return filepath.SkipDir
			}
			return nil
		}

		files = append(files, relPath)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
