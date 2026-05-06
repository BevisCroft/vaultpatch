// Package revert provides functionality to revert a Vault secret path
// to a previous known state captured in a snapshot.
package revert

import (
	"context"
	"fmt"
	"time"

	"github.com/youorg/vaultpatch/internal/audit"
	"github.com/youorg/vaultpatch/internal/snapshot"
	"github.com/youorg/vaultpatch/internal/vault"
)

// Result holds the outcome of a single revert operation.
type Result struct {
	Path    string
	DryRun  bool
	Err     error
	Skipped bool
}

// Reverter applies snapshot state back to Vault.
type Reverter struct {
	client  *vault.Client
	auditor *audit.Auditor
	dryRun  bool
}

// New creates a new Reverter.
func New(client *vault.Client, auditor *audit.Auditor, dryRun bool) *Reverter {
	return &Reverter{client: client, auditor: auditor, dryRun: dryRun}
}

// Apply reverts each path found in the snapshot to its captured values.
// Paths not present in the snapshot are skipped.
func (r *Reverter) Apply(ctx context.Context, snap *snapshot.Snapshot, paths []string) []Result {
	results := make([]Result, 0, len(paths))

	for _, path := range paths {
		res := Result{Path: path, DryRun: r.dryRun}

		data, ok := snap.Secrets[path]
		if !ok {
			res.Skipped = true
			results = append(results, res)
			continue
		}

		if !r.dryRun {
			if err := r.client.WriteSecret(ctx, path, data); err != nil {
				res.Err = fmt.Errorf("write %s: %w", path, err)
				results = append(results, res)
				_ = r.auditor.Record(ctx, audit.Entry{
					Op:        "revert",
					Path:      path,
					DryRun:    false,
					Success:   false,
					Message:   res.Err.Error(),
					Timestamp: time.Now().UTC(),
				})
				continue
			}
			_ = r.auditor.Record(ctx, audit.Entry{
				Op:        "revert",
				Path:      path,
				DryRun:    false,
				Success:   true,
				Timestamp: time.Now().UTC(),
			})
		}

		results = append(results, res)
	}

	return results
}
