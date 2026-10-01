package core

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// alwaysIgnoredDirs are tool/cache directories skipped at any depth.
var alwaysIgnoredDirs = map[string]bool{
	".git":          true,
	"node_modules":  true,
	".next":         true,
	".nuxt":         true,
	".svelte-kit":   true,
	".turbo":        true,
	"__pycache__":   true,
	".pytest_cache": true,
	".mypy_cache":   true,
	".venv":         true,
	".gradle":       true,
	".dart_tool":    true,
	".idea":         true,
	".vscode":       true,
}

// rootIgnoredDirs are common build-output names that are only skipped at the project root,
// because nested directories with these names (e.g. internal/build) are usually real source.
var rootIgnoredDirs = map[string]bool{
	"vendor":   true,
	"bin":      true,
	"dist":     true,
	"build":    true,
	"out":      true,
	".output":  true,
	"coverage": true,
	".cache":   true,
	"target":   true,
	"venv":     true,
	"env":      true,
}

// collectFiles returns the sorted, slash-separated relative paths of files to scan under rootDir.
// Inside a git work tree it uses `git ls-files` so .gitignore is respected; otherwise it walks the tree.
func collectFiles(ctx context.Context, rootDir string, customIgnores []string) ([]string, error) {
	candidates, err := gitListFiles(ctx, rootDir)
	if err != nil {
		candidates, err = walkFiles(rootDir)
		if err != nil {
			return nil, err
		}
	}

	files := make([]string, 0, len(candidates))
	for _, rel := range candidates {
		if !isIgnored(rel, customIgnores) {
			files = append(files, rel)
		}
	}
	sort.Strings(files)
	return files, nil
}

func gitListFiles(ctx context.Context, rootDir string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", rootDir, "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var files []string
	for _, entry := range bytes.Split(out, []byte{0}) {
		rel := string(entry)
		if rel == "" || seen[rel] {
			continue
		}
		seen[rel] = true

		// Skip tracked files deleted from the work tree, and submodule directories.
		info, err := os.Lstat(filepath.Join(rootDir, filepath.FromSlash(rel)))
		if err != nil || info.IsDir() {
			continue
		}
		files = append(files, rel)
	}
	return files, nil
}

func walkFiles(rootDir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(rootDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(rootDir, p)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if alwaysIgnoredDirs[d.Name()] || (!strings.Contains(rel, "/") && rootIgnoredDirs[d.Name()]) {
				return filepath.SkipDir
			}
			return nil
		}
		files = append(files, rel)
		return nil
	})
	return files, err
}

// isIgnored applies the built-in directory ignores and the user's ignore patterns to a
// slash-separated relative path. A pattern without "/" matches any path segment (glob
// syntax allowed, e.g. "*.gen.go" or "tmp"); a pattern containing "/" is anchored at the
// project root and matches that path or anything beneath it (e.g. "docs/generated").
func isIgnored(rel string, patterns []string) bool {
	segments := strings.Split(rel, "/")
	dirs := segments[:len(segments)-1]

	for i, dir := range dirs {
		if alwaysIgnoredDirs[dir] || (i == 0 && rootIgnoredDirs[dir]) {
			return true
		}
	}

	for _, raw := range patterns {
		pattern := strings.Trim(strings.TrimSpace(raw), "/")
		if pattern == "" {
			continue
		}

		if !strings.Contains(pattern, "/") {
			for _, seg := range segments {
				if ok, _ := path.Match(pattern, seg); ok {
					return true
				}
			}
			continue
		}

		depth := strings.Count(pattern, "/") + 1
		if depth > len(segments) {
			continue
		}
		prefix := strings.Join(segments[:depth], "/")
		if ok, _ := path.Match(pattern, prefix); ok {
			return true
		}
	}

	return false
}
