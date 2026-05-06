// Package access provides path-level access control checking for Vault
// operations, allowing callers to verify whether a given token or role
// is permitted to perform an action before attempting it.
package access

import (
	"context"
	"fmt"
	"strings"
)

// Permission represents an allowed operation on a path.
type Permission string

const (
	PermRead   Permission = "read"
	PermWrite  Permission = "write"
	PermDelete Permission = "delete"
	PermList   Permission = "list"
)

// Rule defines what permissions are allowed for a path prefix.
type Rule struct {
	PathPrefix  string
	Permissions []Permission
}

// Result holds the outcome of an access check.
type Result struct {
	Path    string
	Op      Permission
	Allowed bool
	Reason  string
}

// Checker evaluates whether operations are permitted under a set of rules.
type Checker struct {
	rules []Rule
}

// New returns a Checker configured with the provided rules.
func New(rules []Rule) *Checker {
	return &Checker{rules: rules}
}

// Check evaluates whether op is permitted on path.
func (c *Checker) Check(_ context.Context, path string, op Permission) Result {
	for _, r := range c.rules {
		if !strings.HasPrefix(path, r.PathPrefix) {
			continue
		}
		for _, p := range r.Permissions {
			if p == op {
				return Result{Path: path, Op: op, Allowed: true,
					Reason: fmt.Sprintf("matched rule prefix %q", r.PathPrefix)}
			}
		}
		return Result{Path: path, Op: op, Allowed: false,
			Reason: fmt.Sprintf("operation %q not in rule for prefix %q", op, r.PathPrefix)}
	}
	return Result{Path: path, Op: op, Allowed: false,
		Reason: fmt.Sprintf("no rule matched path %q", path)}
}

// CheckAll evaluates multiple (path, op) pairs and returns all results.
func (c *Checker) CheckAll(ctx context.Context, checks []struct {
	Path string
	Op   Permission
}) []Result {
	out := make([]Result, 0, len(checks))
	for _, ch := range checks {
		out = append(out, c.Check(ctx, ch.Path, ch.Op))
	}
	return out
}

// AllAllowed returns true only if every Result in the slice is allowed.
// It is a convenience helper for callers that run CheckAll and need a
// single boolean gate before proceeding with a batch operation.
func AllAllowed(results []Result) bool {
	for _, r := range results {
		if !r.Allowed {
			return false
		}
	}
	return true
}
