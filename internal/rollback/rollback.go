// Package rollback provides functionality to restore Vault secrets
// from a previously captured snapshot, enabling safe rollback of changes.
package rollback

import (
	"context"
	"fmt"

	"github.com/user/vaultpatch/internal/audit"
	"github.com/user/vaultpatch/internal/snapshot"
)

// Writer is the interface required to write secrets back to Vault.
type Writer interface {
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
}

// Applier restores secrets from a snapshot.
type Applier struct {
	client  Writer
	auditor *audit.Auditor
	dryRun  bool
}

// New creates a new rollback Applier.
func New(client Writer, auditor *audit.Auditor, dryRun bool) *Applier {
	return &Applier{
		client:  client,
		auditor: auditor,
		dryRun:  dryRun,
	}
}

// Result holds the outcome of a single rollback operation.
type Result struct {
	Path  string
	Err   error
}

// Apply restores all secrets from the given snapshot.
// If dryRun is true, no writes are performed.
func (a *Applier) Apply(ctx context.Context, snap *snapshot.Snapshot) ([]Result, error) {
	if snap == nil {
		return nil, fmt.Errorf("rollback: snapshot must not be nil")
	}

	results := make([]Result, 0, len(snap.Secrets))

	for path, data := range snap.Secrets {
		var opErr error

		if !a.dryRun {
			opErr = a.client.WriteSecret(ctx, path, data)
		}

		results = append(results, Result{Path: path, Err: opErr})

		if a.auditor != nil {
			msg := ""
			if opErr != nil {
				msg = opErr.Error()
			}
			_ = a.auditor.Record(ctx, "rollback", path, a.dryRun, msg)
		}
	}

	return results, nil
}
