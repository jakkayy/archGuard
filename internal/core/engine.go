package core

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
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

// Rules returns all registered rules sorted by ID.
func (e *Engine) Rules() []Rule {
	list := make([]Rule, 0, len(e.rules))
	for _, r := range e.rules {
		list = append(list, r)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID() < list[j].ID() })
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

	files, err := collectFiles(ctx, absWorkingDir, customIgnores)
	if err != nil {
		return nil, fmt.Errorf("failed to collect project files in %s: %w", absWorkingDir, err)
	}

	scanCtx := NewScanContext(ctx, absWorkingDir, files)

	var allIssues []Issue

	if cfg != nil {
		ruleIDs := make([]string, 0, len(cfg.Rules))
		for id := range cfg.Rules {
			ruleIDs = append(ruleIDs, id)
		}
		sort.Strings(ruleIDs)

		for _, ruleID := range ruleIDs {
			ruleCfg := cfg.Rules[ruleID]
			r, ok := e.rules[ruleID]
			if !ok {
				return nil, e.unknownRuleError(ruleID)
			}

			if !ruleCfg.Enabled {
				continue
			}

			if err := scanCtx.Ctx.Err(); err != nil {
				return nil, fmt.Errorf("scan cancelled: %w", err)
			}

			issues, err := r.Run(scanCtx)
			if err != nil {
				return nil, fmt.Errorf("failed executing rule %s: %w", ruleID, err)
			}

			allIssues = append(allIssues, issues...)
		}
	}

	sortIssues(allIssues)

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

// sortIssues orders issues by file, line, then rule so output is stable across runs.
func sortIssues(issues []Issue) {
	sort.SliceStable(issues, func(i, j int) bool {
		a, b := issues[i], issues[j]
		if a.FilePath != b.FilePath {
			return a.FilePath < b.FilePath
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.RuleID < b.RuleID
	})
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
