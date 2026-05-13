// Package status provides a health-check summary for Vault paths,
// reporting whether each path is readable, missing, or in error.
package status

import (
	"context"
	"fmt"
	"sort"
)

// State represents the health state of a single Vault path.
type State int

const (
	StateOK      State = iota // path exists and is readable
	StateMissing              // path returned 404
	StateError                // unexpected error
)

// Result holds the status check outcome for one path.
type Result struct {
	Path  string
	State State
	Err   error
}

// VaultReader is the subset of the Vault client used by this package.
type VaultReader interface {
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
}

// Checker runs status checks against a set of Vault paths.
type Checker struct {
	client VaultReader
}

// New returns a new Checker backed by the provided VaultReader.
func New(client VaultReader) *Checker {
	return &Checker{client: client}
}

// Check inspects each path and returns a slice of Results sorted by path.
func (c *Checker) Check(ctx context.Context, paths []string) []Result {
	results := make([]Result, 0, len(paths))

	for _, p := range paths {
		_, err := c.client.ReadSecret(ctx, p)
		result := Result{Path: p}

		switch {
		case err == nil:
			result.State = StateOK
		case isMissing(err):
			result.State = StateMissing
		default:
			result.State = StateError
			result.Err = err
		}

		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Path < results[j].Path
	})

	return results
}

// isMissing returns true when err indicates a not-found / 404 condition.
func isMissing(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == fmt.Sprintf("secret not found")
}
