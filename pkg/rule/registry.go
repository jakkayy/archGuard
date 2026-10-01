package rule

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jakkayy/archGuard/pkg/policy"
)

// ParamSpec documents one configuration parameter accepted by a rule.
type ParamSpec struct {
	Name        string
	Type        string
	Description string
}

// ExcludeParam is accepted by every rule: a list of path patterns (see policy.MatchPath)
// whose files the rule skips and whose issues are dropped.
const ExcludeParam = "exclude"

// CommonParams are accepted by every rule in addition to its own Params.
var CommonParams = []ParamSpec{
	{Name: ExcludeParam, Type: "list of strings", Description: "path patterns this rule skips, e.g. ['**/testdata/**', '*_test.go']"},
}

// Factory constructs a rule from its validated parameters and resolved severity.
type Factory func(params Params, severity policy.Severity) (policy.Rule, error)

// Definition describes a rule that can be enabled from archguard.yaml.
type Definition struct {
	ID              string
	Name            string
	Description     string
	DefaultSeverity policy.Severity
	Params          []ParamSpec
	Factory         Factory
}

// Build validates params against the definition and constructs the rule.
// An empty severity falls back to DefaultSeverity.
func (d Definition) Build(params map[string]any, severity policy.Severity) (policy.Rule, error) {
	allowed := make(map[string]bool, len(d.Params)+len(CommonParams))
	for _, p := range append(append([]ParamSpec{}, d.Params...), CommonParams...) {
		allowed[p.Name] = true
	}
	for key := range params {
		if !allowed[key] {
			return nil, fmt.Errorf("rule %q: unknown parameter %q (allowed: %s)", d.ID, key, d.paramNames())
		}
	}

	p := Params{ruleID: d.ID, values: params}
	exclude, err := p.StringSlice(ExcludeParam)
	if err != nil {
		return nil, fmt.Errorf("rule %q: %w", d.ID, err)
	}
	for _, pattern := range exclude {
		if err := policy.ValidatePathPattern(pattern); err != nil {
			return nil, fmt.Errorf("rule %q: %w", d.ID, err)
		}
	}

	if severity == "" {
		severity = d.DefaultSeverity
	}
	r, err := d.Factory(p, severity)
	if err != nil {
		return nil, fmt.Errorf("rule %q: %w", d.ID, err)
	}
	if len(exclude) > 0 {
		r = &excludingRule{Rule: r, patterns: exclude}
	}
	return r, nil
}

func (d Definition) paramNames() string {
	names := make([]string, 0, len(d.Params)+len(CommonParams))
	for _, p := range d.Params {
		names = append(names, p.Name)
	}
	for _, p := range CommonParams {
		names = append(names, p.Name)
	}
	return strings.Join(names, ", ")
}

// excludingRule hides excluded files from the wrapped rule and drops issues reported on them.
type excludingRule struct {
	policy.Rule
	patterns []string
}

func (r *excludingRule) excluded(rel string) bool {
	for _, p := range r.patterns {
		if policy.MatchPath(p, rel) {
			return true
		}
	}
	return false
}

// Run executes the wrapped rule against the non-excluded files only.
func (r *excludingRule) Run(ctx *policy.ScanContext) ([]policy.Issue, error) {
	scoped := *ctx
	scoped.Files = make([]string, 0, len(ctx.Files))
	for _, f := range ctx.Files {
		if !r.excluded(f) {
			scoped.Files = append(scoped.Files, f)
		}
	}

	issues, err := r.Rule.Run(&scoped)
	if err != nil {
		return nil, err
	}

	kept := issues[:0]
	for _, issue := range issues {
		if !r.excluded(issue.FilePath) {
			kept = append(kept, issue)
		}
	}
	return kept, nil
}

// Registry holds rule definitions keyed by rule ID.
type Registry struct {
	defs map[string]Definition
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{defs: make(map[string]Definition)}
}

// Register adds a definition, rejecting duplicate or incomplete entries.
func (r *Registry) Register(d Definition) error {
	if d.ID == "" || d.Factory == nil {
		return fmt.Errorf("rule definition must have an ID and a Factory")
	}
	if _, exists := r.defs[d.ID]; exists {
		return fmt.Errorf("rule %q is already registered", d.ID)
	}
	r.defs[d.ID] = d
	return nil
}

// Lookup returns the definition for id.
func (r *Registry) Lookup(id string) (Definition, bool) {
	d, ok := r.defs[id]
	return d, ok
}

// Definitions returns all definitions sorted by ID.
func (r *Registry) Definitions() []Definition {
	list := make([]Definition, 0, len(r.defs))
	for _, d := range r.defs {
		list = append(list, d)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	return list
}

// Builtins returns a registry containing every rule shipped with ArchGuard.
func Builtins() *Registry {
	reg := NewRegistry()
	for _, d := range []Definition{
		{
			ID:              "file-naming",
			Name:            "File Naming Convention",
			Description:     "Validates that project filenames conform to the specified regex pattern",
			DefaultSeverity: policy.SeverityWarning,
			Params: []ParamSpec{
				{Name: "pattern", Type: "string", Description: "regex every file name must match (default " + DefaultFileNamingPattern + ")"},
			},
			Factory: func(p Params, sev policy.Severity) (policy.Rule, error) {
				pattern, err := p.String("pattern")
				if err != nil {
					return nil, err
				}
				return NewFileNamingRule(pattern, sev)
			},
		},
		{
			ID:              "no-secrets",
			Name:            "No Hardcoded Secrets Check",
			Description:     "Scans project files for hardcoded API keys, private keys, and tokens",
			DefaultSeverity: policy.SeverityError,
			Factory: func(_ Params, sev policy.Severity) (policy.Rule, error) {
				return NewNoSecretsRule(sev), nil
			},
		},
		{
			ID:              "openapi-exists",
			Name:            "OpenAPI Spec Existence Check",
			Description:     "Validates that the required OpenAPI specification file exists in the workspace",
			DefaultSeverity: policy.SeverityError,
			Params: []ParamSpec{
				{Name: "path", Type: "string", Description: "spec file path relative to the project root (default " + DefaultOpenAPIPath + ")"},
			},
			Factory: func(p Params, sev policy.Severity) (policy.Rule, error) {
				path, err := p.String("path")
				if err != nil {
					return nil, err
				}
				return NewOpenAPIExistsRule(path, sev), nil
			},
		},
		{
			ID:              "required-files",
			Name:            "Required Files Existence Check",
			Description:     "Validates that mandatory files exist in the project workspace",
			DefaultSeverity: policy.SeverityError,
			Params: []ParamSpec{
				{Name: "files", Type: "list of strings", Description: "paths that must exist, e.g. ['README.md', '.gitignore'] (default ['README.md'])"},
			},
			Factory: func(p Params, sev policy.Severity) (policy.Rule, error) {
				files, err := p.StringSlice("files")
				if err != nil {
					return nil, err
				}
				return NewRequiredFilesRule(files, sev), nil
			},
		},
	} {
		if err := reg.Register(d); err != nil {
			panic(err) // programming error in the built-in table
		}
	}
	return reg
}

// Params gives typed access to a rule's raw YAML parameters.
type Params struct {
	ruleID string
	values map[string]any
}

// String returns the string parameter key, or "" if it is absent.
func (p Params) String(key string) (string, error) {
	raw, ok := p.values[key]
	if !ok || raw == nil {
		return "", nil
	}
	s, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("parameter %q must be a string, got %T", key, raw)
	}
	return s, nil
}

// StringSlice returns the list-of-strings parameter key, or nil if it is absent.
func (p Params) StringSlice(key string) ([]string, error) {
	raw, ok := p.values[key]
	if !ok || raw == nil {
		return nil, nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("parameter %q must be a list of strings, got %T", key, raw)
	}
	out := make([]string, 0, len(items))
	for i, item := range items {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("parameter %q[%d] must be a string, got %T", key, i, item)
		}
		out = append(out, s)
	}
	return out, nil
}
