// Package prune removes secret paths from Vault that match a given age
// threshold and have not been accessed within that window.
package prune

import (
	"context"
	"fmt"
	"time"
)

// VaultClient is the subset of vault.Client used by Prune.
type VaultClient interface {
	ListSecrets(ctx context.Context, path string) ([]string, error)
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
	DeleteSecret(ctx context.Context, path string) error
}

// Auditor records what Prune does.
type Auditor interface {
	Record(ctx context.Context, op, path string, dryRun bool, err error) error
}

// Result holds the outcome for a single path.
type Result struct {
	Path    string
	Pruned  bool
	DryRun  bool
	Skipped bool
	Err     error
}

// Pruner walks a Vault path and removes secrets whose "updated_at" metadata
// is older than MaxAge.
type Pruner struct {
	client VaultClient
	auditor Auditor
	MaxAge time.Duration
	DryRun bool
}

// New creates a Pruner.
func New(client VaultClient, auditor Auditor, maxAge time.Duration, dryRun bool) *Pruner {
	return &Pruner{
		client:  client,
		auditor: auditor,
		MaxAge:  maxAge,
		DryRun:  dryRun,
	}
}

// Run lists all secrets under root and prunes those that are stale.
func (p *Pruner) Run(ctx context.Context, root string) ([]Result, error) {
	paths, err := p.client.ListSecrets(ctx, root)
	if err != nil {
		return nil, fmt.Errorf("prune: list %q: %w", root, err)
	}

	var results []Result
	cutoff := time.Now().UTC().Add(-p.MaxAge)

	for _, path := range paths {
		secret, err := p.client.ReadSecret(ctx, path)
		if err != nil {
			results = append(results, Result{Path: path, Err: err})
			_ = p.auditor.Record(ctx, "prune", path, p.DryRun, err)
			continue
		}

		updatedRaw, ok := secret["updated_at"]
		if !ok {
			results = append(results, Result{Path: path, Skipped: true})
			continue
		}

		updatedAt, err := time.Parse(time.RFC3339, updatedRaw)
		if err != nil {
			results = append(results, Result{Path: path, Skipped: true})
			continue
		}

		if updatedAt.After(cutoff) {
			results = append(results, Result{Path: path, Skipped: true})
			continue
		}

		var delErr error
		if !p.DryRun {
			delErr = p.client.DeleteSecret(ctx, path)
		}
		_ = p.auditor.Record(ctx, "prune", path, p.DryRun, delErr)
		results = append(results, Result{Path: path, Pruned: true, DryRun: p.DryRun, Err: delErr})
	}

	return results, nil
}
