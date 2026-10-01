package policy_test

import (
	"testing"

	"github.com/jakkayy/archGuard/pkg/policy"
)

func TestMatchPath(t *testing.T) {
	tests := []struct {
		pattern, rel string
		want         bool
	}{
		{"testdata", "testdata/key.pem", true},
		{"testdata", "internal/backend/testdata/certs/ca.key", true},
		{"*_test.go", "internal/lang/funcs/crypto_test.go", true},
		{"*_test.go", "internal/lang/funcs/crypto.go", false},
		{"docs/generated", "docs/generated/a.md", true},
		{"docs/generated", "other/docs/generated/a.md", false},
		{"/docs/generated/", "docs/generated/a.md", true},
		{"**/testdata/**", "testdata/key.pem", true},
		{"**/testdata/**", "a/b/testdata/c/key.pem", true},
		{"**/testdata/**", "a/b/c.go", false},
		{"internal/**/fixtures", "internal/x/y/fixtures/f.json", true},
		{"internal/**/fixtures", "internal/fixtures/f.json", true},
		{"internal/**/fixtures", "pkg/fixtures/f.json", false},
		{"docs_src/**/*.py", "docs_src/security/tutorial004.py", true},
		{"", "anything", false},
	}
	for _, tt := range tests {
		if got := policy.MatchPath(tt.pattern, tt.rel); got != tt.want {
			t.Errorf("MatchPath(%q, %q) = %v, want %v", tt.pattern, tt.rel, got, tt.want)
		}
	}
}

func TestValidatePathPattern(t *testing.T) {
	if err := policy.ValidatePathPattern("**/testdata/**"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := policy.ValidatePathPattern("bad/[pattern"); err == nil {
		t.Error("expected error for malformed pattern")
	}
}
