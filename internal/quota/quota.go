// Package quota provides utilities for checking and enforcing secret count
// limits across Vault paths.
package quota

import (
	"context"
	"fmt"
)

// SecretLister can list secret keys under a path.
type SecretLister interface {
	ListSecrets(ctx context.Context, path string) ([]string, error)
}

// Rule defines a quota limit for a path prefix.
type Rule struct {
	Path  string
	Limit int
}

// Result holds the outcome of a quota check for a single path.
type Result struct {
	Path    string
	Limit   int
	Current int
	Exceeds bool
	Err     error
}

// Checker evaluates quota rules against live Vault paths.
type Checker struct {
	client SecretLister
	rules  []Rule
}

// New creates a Checker with the given Vault client and quota rules.
func New(client SecretLister, rules []Rule) *Checker {
	return &Checker{client: client, rules: rules}
}

// Check evaluates all rules and returns one Result per rule.
func (c *Checker) Check(ctx context.Context) []Result {
	results := make([]Result, 0, len(c.rules))
	for _, rule := range c.rules {
		res := Result{Path: rule.Path, Limit: rule.Limit}
		keys, err := c.client.ListSecrets(ctx, rule.Path)
		if err != nil {
			res.Err = fmt.Errorf("list %q: %w", rule.Path, err)
			results = append(results, res)
			continue
		}
		res.Current = len(keys)
		res.Exceeds = res.Current > rule.Limit
		results = append(results, res)
	}
	return results
}

// AnyExceeds returns true if any result exceeds its quota.
func AnyExceeds(results []Result) bool {
	for _, r := range results {
		if r.Exceeds {
			return true
		}
	}
	return false
}
