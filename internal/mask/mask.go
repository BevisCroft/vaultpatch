// Package mask provides utilities for selectively masking secret values
// based on configurable path and key patterns before display or export.
package mask

import (
	"regexp"
	"strings"
)

// Rule describes a single masking rule.
type Rule struct {
	// PathPattern is a glob-style prefix matched against the secret path.
	PathPattern string
	// KeyPattern is a regexp matched against the secret key name.
	KeyPattern *regexp.Regexp
	// Replacement is the string used in place of the real value.
	Replacement string
}

// Masker applies a set of Rules to secret maps.
type Masker struct {
	rules []Rule
}

// DefaultRules returns a sensible set of built-in masking rules that cover
// common sensitive key names (passwords, tokens, keys, secrets).
func DefaultRules() []Rule {
	pattern := regexp.MustCompile(`(?i)(password|passwd|secret|token|api[_-]?key|private[_-]?key|auth)`)
	return []Rule{
		{PathPattern: "*", KeyPattern: pattern, Replacement: "***"},
	}
}

// New creates a Masker with the provided rules. If rules is empty the
// DefaultRules are used.
func New(rules []Rule) *Masker {
	if len(rules) == 0 {
		rules = DefaultRules()
	}
	return &Masker{rules: rules}
}

// Apply returns a copy of secrets with sensitive values replaced according to
// the configured rules. The original map is never mutated.
func (m *Masker) Apply(path string, secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		out[k] = v
	}
	for _, rule := range m.rules {
		if !matchesPath(rule.PathPattern, path) {
			continue
		}
		for k := range out {
			if rule.KeyPattern.MatchString(k) {
				out[k] = rule.Replacement
			}
		}
	}
	return out
}

// matchesPath returns true when pattern (glob prefix) matches path.
func matchesPath(pattern, path string) bool {
	if pattern == "*" {
		return true
	}
	return strings.HasPrefix(path, strings.TrimSuffix(pattern, "*"))
}
