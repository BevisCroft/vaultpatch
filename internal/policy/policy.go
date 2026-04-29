// Package policy provides utilities for checking Vault policy compliance
// against a set of secret paths and operations.
package policy

import (
	"fmt"
	"strings"
)

// Op represents a Vault operation type.
type Op string

const (
	OpRead   Op = "read"
	OpWrite  Op = "write"
	OpDelete Op = "delete"
	OpList   Op = "list"
)

// Rule defines an allowed operation on a path pattern.
type Rule struct {
	PathPattern string
	AllowedOps  []Op
}

// Policy holds a named collection of rules.
type Policy struct {
	Name  string
	Rules []Rule
}

// Violation describes a path/op pair that is not permitted by the policy.
type Violation struct {
	Path string
	Op   Op
}

func (v Violation) Error() string {
	return fmt.Sprintf("operation %q not allowed on path %q", v.Op, v.Path)
}

// Check evaluates whether all (path, op) pairs are permitted by p.
// It returns a slice of Violations for any that are not covered.
func (p *Policy) Check(requests map[string]Op) []Violation {
	var violations []Violation
	for path, op := range requests {
		if !p.allowed(path, op) {
			violations = append(violations, Violation{Path: path, Op: op})
		}
	}
	return violations
}

// allowed returns true when at least one rule covers path and permits op.
func (p *Policy) allowed(path string, op Op) bool {
	for _, rule := range p.Rules {
		if matchPattern(rule.PathPattern, path) {
			for _, allowed := range rule.AllowedOps {
				if allowed == op {
					return true
				}
			}
		}
	}
	return false
}

// matchPattern supports a single trailing '*' wildcard.
func matchPattern(pattern, path string) bool {
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(path, strings.TrimSuffix(pattern, "*"))
	}
	return pattern == path
}
