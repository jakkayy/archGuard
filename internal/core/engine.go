package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/archguard/archguard/internal/config"
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
			if !ruleCfg.Enabled {
				continue
			}

			r, ok := e.rules[ruleID]
			if !ok {
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
