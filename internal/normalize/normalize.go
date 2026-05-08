// Package normalize provides utilities for standardizing Vault secret keys
// across environments, enforcing consistent casing and formatting conventions.
package normalize

import (
	"strings"
	"unicode"
)

// Rule defines a normalization strategy for secret keys.
type Rule string

const (
	// RuleLowercase converts all keys to lowercase.
	RuleLowercase Rule = "lowercase"
	// RuleUppercase converts all keys to UPPERCASE.
	RuleUppercase Rule = "uppercase"
	// RuleSnakeCase converts keys to snake_case.
	RuleSnakeCase Rule = "snake_case"
	// RuleKebabCase converts keys to kebab-case.
	RuleKebabCase Rule = "kebab-case"
)

// Result holds the outcome of normalizing a single key.
type Result struct {
	Original   string
	Normalized string
	Changed    bool
}

// Normalizer applies a normalization rule to secret keys.
type Normalizer struct {
	rule Rule
}

// New creates a Normalizer with the given rule.
func New(rule Rule) (*Normalizer, error) {
	switch rule {
	case RuleLowercase, RuleUppercase, RuleSnakeCase, RuleKebabCase:
		return &Normalizer{rule: rule}, nil
	default:
		return nil, fmt.Errorf("normalize: unknown rule %q", rule)
	}
}

// Apply normalizes the keys of the provided secrets map and returns a slice
// of Results describing each transformation. The original map is not mutated.
func (n *Normalizer) Apply(secrets map[string]string) (map[string]string, []Result) {
	out := make(map[string]string, len(secrets))
	results := make([]Result, 0, len(secrets))

	for k, v := range secrets {
		norm := n.normalize(k)
		results = append(results, Result{
			Original:   k,
			Normalized: norm,
			Changed:    norm != k,
		})
		out[norm] = v
	}
	return out, results
}

func (n *Normalizer) normalize(key string) string {
	switch n.rule {
	case RuleLowercase:
		return strings.ToLower(key)
	case RuleUppercase:
		return strings.ToUpper(key)
	case RuleSnakeCase:
		return toSnakeCase(key)
	case RuleKebabCase:
		return strings.ReplaceAll(toSnakeCase(key), "_", "-")
	}
	return key
}

// toSnakeCase converts a string to snake_case, splitting on spaces, hyphens,
// and camelCase boundaries.
func toSnakeCase(s string) string {
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			prev := rune(s[i-1])
			if unicode.IsLower(prev) || prev == '_' {
				b.WriteRune('_')
			}
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}
