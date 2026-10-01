package core

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func writeTree(t *testing.T, root string, files ...string) {
	t.Helper()
	for _, f := range files {
		full := filepath.Join(root, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestIsIgnored(t *testing.T) {
	tests := []struct {
		rel      string
		patterns []string
		want     bool
	}{
		{"node_modules/pkg/index.js", nil, true},
		{"web/node_modules/pkg/index.js", nil, true},
		{"build/out.js", nil, true},
		{"internal/build/build.go", nil, false},
		{"pkg/env/env.go", nil, false},
		{"main.go", nil, false},
		{"tmp/a.go", []string{"tmp"}, true},
		{"src/tmp/a.go", []string{"tmp"}, true},
		{"api/types.gen.go", []string{"*.gen.go"}, true},
		{"docs/generated/a.md", []string{"docs/generated"}, true},
		{"other/docs/generated/a.md", []string{"docs/generated"}, false},
		{"docs/generated/a.md", []string{"/docs/generated/"}, true},
	}
	for _, tt := range tests {
		if got := isIgnored(tt.rel, tt.patterns); got != tt.want {
			t.Errorf("isIgnored(%q, %v) = %v, want %v", tt.rel, tt.patterns, got, tt.want)
		}
	}
}

func TestCollectFiles_WalkKeepsNestedBuildDirs(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, "main.go", "internal/build/build.go", "build/artifact.js", "node_modules/x/index.js")

	got, err := collectFiles(context.Background(), root, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"internal/build/build.go", "main.go"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("collectFiles = %v, want %v", got, want)
	}
}

func TestCollectFiles_RespectsGitignore(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v: %s", err, out)
	}
	writeTree(t, root, "main.go", "secrets.local", "logs/app.log")
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.local\nlogs/\n"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := collectFiles(context.Background(), root, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{".gitignore", "main.go"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("collectFiles = %v, want %v", got, want)
	}
}
