package policy

import (
	"fmt"
	"path"
	"strings"
)

// MatchPath reports whether the slash-separated relative path rel matches pattern.
//
//   - A pattern without "/" is matched against every path segment, so "testdata"
//     matches any directory named testdata and "*_test.go" matches any test file.
//   - A pattern containing "/" is anchored at the project root and matches that path
//     or anything beneath it, so "docs/generated" matches "docs/generated/a.md".
//   - Within an anchored pattern, a "**" segment matches zero or more segments,
//     e.g. "**/testdata/**" or "internal/**/fixtures".
//
// Segments use path.Match glob syntax.
func MatchPath(pattern, rel string) bool {
	pattern = strings.Trim(strings.TrimSpace(pattern), "/")
	if pattern == "" {
		return false
	}
	segments := strings.Split(rel, "/")

	if !strings.Contains(pattern, "/") {
		for _, seg := range segments {
			if ok, _ := path.Match(pattern, seg); ok {
				return true
			}
		}
		return false
	}

	return matchPrefix(strings.Split(pattern, "/"), segments)
}

// matchPrefix reports whether pat matches a leading run of segs.
func matchPrefix(pat, segs []string) bool {
	if len(pat) == 0 {
		return true
	}
	if pat[0] == "**" {
		for i := 0; i <= len(segs); i++ {
			if matchPrefix(pat[1:], segs[i:]) {
				return true
			}
		}
		return false
	}
	if len(segs) == 0 {
		return false
	}
	if ok, _ := path.Match(pat[0], segs[0]); !ok {
		return false
	}
	return matchPrefix(pat[1:], segs[1:])
}

// ValidatePathPattern returns an error if pattern contains malformed glob syntax.
func ValidatePathPattern(pattern string) error {
	for _, seg := range strings.Split(strings.Trim(pattern, "/"), "/") {
		if seg == "**" {
			continue
		}
		if _, err := path.Match(seg, ""); err != nil {
			return fmt.Errorf("invalid path pattern %q: %w", pattern, err)
		}
	}
	return nil
}
