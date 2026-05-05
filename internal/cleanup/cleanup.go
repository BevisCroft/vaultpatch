// Package cleanup identifies and removes stale Vault secrets based on
// configurable age thresholds and optional dry-run protection.
package cleanup

import (
	"context"
	"fmt"
	"time"

	"github.com/your-org/vaultpatch/internal/audit"
)

// VaultClient is the subset of vault operations required by the cleaner.
type VaultClient interface {
	ListSecrets(ctx context.Context, path string) ([]string, error)
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
	DeleteSecret(ctx context.Context, path string) error
}

// Result holds the outcome of a single cleanup candidate.
type Result struct {
	Path    string
	Age     time.Duration
	Deleted bool
	DryRun  bool
	Err     error
}

// Cleaner scans a Vault path prefix and removes secrets older than MaxAge.
type Cleaner struct {
	client  VaultClient
	auditor *audit.Auditor
	MaxAge  time.Duration
	DryRun  bool
}

// New returns a Cleaner configured with the provided client and auditor.
func New(client VaultClient, aud *audit.Auditor, maxAge time.Duration, dryRun bool) *Cleaner {
	return &Cleaner{
		client:  client,
		auditor: aud,
		MaxAge:  maxAge,
		DryRun:  dryRun,
	}
}

// Run lists secrets under prefix and deletes those whose "updated_at" metadata
// indicates they are older than MaxAge. Returns one Result per candidate.
func (c *Cleaner) Run(ctx context.Context, prefix string) ([]Result, error) {
	paths, err := c.client.ListSecrets(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("cleanup: list %q: %w", prefix, err)
	}

	var results []Result
	for _, p := range paths {
		data, err := c.client.ReadSecret(ctx, p)
		if err != nil {
			results = append(results, Result{Path: p, Err: err})
			continue
		}

		updatedRaw, ok := data["updated_at"]
		if !ok {
			continue // no timestamp metadata — skip
		}

		updatedAt, err := time.Parse(time.RFC3339, updatedRaw)
		if err != nil {
			results = append(results, Result{Path: p, Err: fmt.Errorf("parse updated_at: %w", err)})
			continue
		}

		age := time.Since(updatedAt)
		if age < c.MaxAge {
			continue
		}

		res := Result{Path: p, Age: age, DryRun: c.DryRun}
		if !c.DryRun {
			if delErr := c.client.DeleteSecret(ctx, p); delErr != nil {
				res.Err = delErr
			} else {
				res.Deleted = true
			}
		}
		_ = c.auditor.Record(ctx, audit.Entry{
			Op:     "cleanup",
			Path:   p,
			DryRun: c.DryRun,
			Err:    res.Err,
		})
		results = append(results, res)
	}
	return results, nil
}
