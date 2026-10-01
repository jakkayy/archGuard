package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jakkayy/archGuard/internal/config"
)

func TestRenderConfig_RoundTripsThroughLoader(t *testing.T) {
	a := defaultAnswers()
	a.Category = "Backend REST API"
	a.Framework = "Go"
	a.OpenAPIEnabled = true
	a.OpenAPIPath = `api/it's "quoted".yaml`
	a.NamingPattern = flexibleNamingPattern

	content, err := renderConfig(a)
	if err != nil {
		t.Fatalf("renderConfig: %v", err)
	}

	path := filepath.Join(t.TempDir(), "archguard.yaml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("generated config failed to load: %v\n%s", err, content)
	}

	if got := cfg.Rules["openapi-exists"].Params["path"]; got != a.OpenAPIPath {
		t.Errorf("openapi path = %v, want %q", got, a.OpenAPIPath)
	}
	if got := cfg.Rules["file-naming"].Params["pattern"]; got != flexibleNamingPattern {
		t.Errorf("pattern = %v, want %q", got, flexibleNamingPattern)
	}
	for _, id := range []string{"no-secrets", "required-files"} {
		if !cfg.Rules[id].Enabled {
			t.Errorf("expected %s to be enabled by default", id)
		}
	}
	if _, err := buildEngine(cfg); err != nil {
		t.Errorf("generated config rejected by engine: %v", err)
	}
}
