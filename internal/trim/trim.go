// Package trim provides functionality to remove secrets from a Vault path
// whose keys match a given set of patterns, with dry-run support.
package trim

import (
	"context"
	"fmt"
	"path"
	"strings"
)

// Writer is the subset of the Vault client used by Trim.
type Writer interface {
	ReadSecret(ctx context.Context, mount, secretPath string) (map[string]string, error)
	WriteSecret(ctx context.Context, mount, secretPath string, data map[string]string) error
}

// Result holds the outcome of a single trim operation.
type Result struct {
	Path    string
	Removed []string
	DryRun  bool
	Err     error
}

// Trimmer removes matching keys from Vault secrets.
type Trimmer struct {
	client Writer
	mount  string
	dryRun bool
}

// New creates a new Trimmer.
func New(client Writer, mount string, dryRun bool) *Trimmer {
	return &Trimmer{client: client, mount: mount, dryRun: dryRun}
}

// Apply reads the secret at secretPath, removes any keys that match one of
// the provided patterns (glob-style), and writes the result back unless
// dry-run is enabled.
func (t *Trimmer) Apply(ctx context.Context, secretPath string, patterns []string) Result {
	res := Result{Path: secretPath, DryRun: t.dryRun}

	data, err := t.client.ReadSecret(ctx, t.mount, secretPath)
	if err != nil {
		res.Err = fmt.Errorf("read %s: %w", secretPath, err)
		return res
	}

	updated := make(map[string]string, len(data))
	for k, v := range data {
		if matchesAny(k, patterns) {
			res.Removed = append(res.Removed, k)
		} else {
			updated[k] = v
		}
	}

	if len(res.Removed) == 0 {
		return res
	}

	if !t.dryRun {
		if err := t.client.WriteSecret(ctx, t.mount, secretPath, updated); err != nil {
			res.Err = fmt.Errorf("write %s: %w", secretPath, err)
		}
	}
	return res
}

// matchesAny reports whether key matches any of the glob patterns.
func matchesAny(key string, patterns []string) bool {
	for _, p := range patterns {
		matched, err := path.Match(strings.ToLower(p), strings.ToLower(key))
		if err == nil && matched {
			return true
		}
	}
	return false
}
