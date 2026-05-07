// Package lint checks Vault secret paths and values against a set of
// configurable rules, reporting style and structural violations.
package lint

import (
	"fmt"
	"regexp"
	"strings"
)

// Severity indicates how serious a lint finding is.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Finding represents a single lint violation.
type Finding struct {
	Path     string
	Key      string
	Message  string
	Severity Severity
}

// Rule is a single lint check applied to each key/value pair.
type Rule struct {
	Name    string
	Check   func(path, key, value string) *Finding
}

// DefaultRules returns the built-in set of lint rules.
func DefaultRules() []Rule {
	return []Rule{
		{
			Name: "no-uppercase-key",
			Check: func(path, key, _ string) *Finding {
				if key != strings.ToLower(key) {
					return &Finding{
						Path:     path,
						Key:      key,
						Message:  "key contains uppercase letters; prefer snake_case",
						Severity: SeverityWarning,
					}
				}
				return nil
			},
		},
		{
			Name: "no-empty-value",
			Check: func(path, key, value string) *Finding {
				if strings.TrimSpace(value) == "" {
					return &Finding{
						Path:     path,
						Key:      key,
						Message:  "value is empty or whitespace-only",
						Severity: SeverityError,
					}
				}
				return nil
			},
		},
		{
			Name: "no-whitespace-in-key",
			Check: func(path, key, _ string) *Finding {
				if strings.ContainsAny(key, " \t") {
					return &Finding{
						Path:     path,
						Key:      key,
						Message:  "key contains whitespace",
						Severity: SeverityError,
					}
				}
				return nil
			},
		},
		{
			Name: "no-plaintext-secret-name",
			Check: func(path, key, _ string) *Finding {
				pattern := regexp.MustCompile(`(?i)(password|passwd|secret|token|apikey|api_key)$`)
				if pattern.MatchString(key) {
					return &Finding{
						Path:     path,
						Key:      key,
						Message:  fmt.Sprintf("key %q looks like a sensitive credential; ensure value is not stored in plaintext", key),
						Severity: SeverityWarning,
					}
				}
				return nil
			},
		},
	}
}

// Linter runs a set of rules against a map of secrets.
type Linter struct {
	rules []Rule
}

// New creates a Linter with the provided rules.
func New(rules []Rule) *Linter {
	return &Linter{rules: rules}
}

// Run applies all rules to the given path and its key/value pairs,
// returning every finding discovered.
func (l *Linter) Run(path string, secrets map[string]string) []Finding {
	var findings []Finding
	for k, v := range secrets {
		for _, rule := range l.rules {
			if f := rule.Check(path, k, v); f != nil {
				findings = append(findings, *f)
			}
		}
	}
	return findings
}
