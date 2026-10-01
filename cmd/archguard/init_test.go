package main

import (
	"os"
	"path/filepath"
	"strings"
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

// scriptedPrompter answers each Choose with the first choice that starts with the next scripted prefix.
type scriptedPrompter struct {
	t       *testing.T
	choices []string
	input   string
}

func (s *scriptedPrompter) Choose(prompt string, choices []string) (string, error) {
	if len(s.choices) == 0 {
		s.t.Fatalf("unexpected prompt %q", prompt)
	}
	prefix := s.choices[0]
	s.choices = s.choices[1:]
	for _, c := range choices {
		if strings.HasPrefix(c, prefix) {
			return c, nil
		}
	}
	s.t.Fatalf("no choice with prefix %q for prompt %q", prefix, prompt)
	return "", nil
}

func (s *scriptedPrompter) Input(string, string) (string, error) { return s.input, nil }

func TestRunWizard(t *testing.T) {
	tests := []struct {
		name        string
		choices     []string
		input       string
		wantPattern string
		wantNaming  bool
		wantOpenAPI bool
		wantPath    string
	}{
		{"frontend flexible", []string{"Frontend", "Next.js", "Flexible"}, "", flexibleNamingPattern, true, false, defaultOpenAPIPath},
		{"backend strict with openapi", []string{"Backend", "Go", "Yes", "Strict"}, " api/spec.yaml ", strictNamingPattern, true, true, "api/spec.yaml"},
		{"backend openapi blank path keeps default", []string{"Backend", "Python", "Yes", "Strict"}, "  ", strictNamingPattern, true, true, defaultOpenAPIPath},
		{"full-stack naming disabled", []string{"Full-Stack", "React", "Node.js", "No", "Disabled"}, "", defaultNamingPattern, false, false, defaultOpenAPIPath},
		{"library", []string{"Library", "Go", "Strict"}, "", strictNamingPattern, true, false, defaultOpenAPIPath},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &scriptedPrompter{t: t, choices: tt.choices, input: tt.input}
			a, err := runWizard(p)
			if err != nil {
				t.Fatal(err)
			}
			if len(p.choices) != 0 {
				t.Errorf("wizard skipped prompts; unused answers %v", p.choices)
			}
			if a.NamingEnabled != tt.wantNaming || (tt.wantNaming && a.NamingPattern != tt.wantPattern) {
				t.Errorf("naming = %v %q, want %v %q", a.NamingEnabled, a.NamingPattern, tt.wantNaming, tt.wantPattern)
			}
			if a.OpenAPIEnabled != tt.wantOpenAPI || a.OpenAPIPath != tt.wantPath {
				t.Errorf("openapi = %v %q, want %v %q", a.OpenAPIEnabled, a.OpenAPIPath, tt.wantOpenAPI, tt.wantPath)
			}
		})
	}
}
