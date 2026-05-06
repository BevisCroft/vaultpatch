// Package redact provides utilities for selectively masking or removing
// sensitive keys from Vault secret maps before display or export.
package redact

import (
	"strings"
)

// Rule defines a redaction rule applied to secret keys.
type Rule struct {
	// Pattern is a case-insensitive substring or exact key name to match.
	Pattern string
	// Replacement is the value substituted for matched keys. Defaults to "***REDACTED***".
	Replacement string
}

// Redactor applies a set of rules to secret maps.
type Redactor struct {
	rules []Rule
}

// DefaultRules returns a sensible set of built-in redaction rules.
func DefaultRules() []Rule {
	return []Rule{
		{Pattern: "password"},
		{Pattern: "secret"},
		{Pattern: "token"},
		{Pattern: "apikey"},
		{Pattern: "api_key"},
		{Pattern: "private_key"},
		{Pattern: "credential"},
	}
}

// New creates a Redactor with the provided rules. If rules is empty,
// DefaultRules are used.
func New(rules []Rule) *Redactor {
	if len(rules) == 0 {
		rules = DefaultRules()
	}
	for i, r := range rules {
		if r.Replacement == "" {
			rules[i].Replacement = "***REDACTED***"
		}
	}
	return &Redactor{rules: rules}
}

// Apply returns a copy of secrets with matched keys replaced by their
// rule's Replacement value. The original map is never mutated.
func (r *Redactor) Apply(secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		out[k] = v
	}
	for k := range out {
		if repl, ok := r.match(k); ok {
			out[k] = repl
		}
	}
	return out
}

// match returns the replacement string and true if key matches any rule.
func (r *Redactor) match(key string) (string, bool) {
	lower := strings.ToLower(key)
	for _, rule := range r.rules {
		if strings.Contains(lower, strings.ToLower(rule.Pattern)) {
			return rule.Replacement, true
		}
	}
	return "", false
}
