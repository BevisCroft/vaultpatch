// Package patch applies computed diffs to a Vault instance.
package patch

import (
	"context"
	"fmt"

	"github.com/example/vaultpatch/internal/diff"
	"github.com/example/vaultpatch/internal/vault"
)

// Result holds the outcome of a single patch operation.
type Result struct {
	Path    string
	Key     string
	Op      diff.Op
	Success bool
	Err     error
}

// Applier applies a set of diff entries to a Vault client.
type Applier struct {
	client *vault.Client
	dryRun bool
}

// NewApplier creates a new Applier.
func NewApplier(client *vault.Client, dryRun bool) *Applier {
	return &Applier{client: client, dryRun: dryRun}
}

// Apply iterates over entries and writes/deletes secrets accordingly.
// It returns a slice of Results, one per entry.
func (a *Applier) Apply(ctx context.Context, entries []diff.Entry) []Result {
	results := make([]Result, 0, len(entries))

	for _, e := range entries {
		r := Result{Path: e.Path, Key: e.Key, Op: e.Op}

		if e.Op == diff.OpNone {
			continue
		}

		if a.dryRun {
			r.Success = true
			results = append(results, r)
			continue
		}

		var err error
		switch e.Op {
		case diff.OpAdd, diff.OpUpdate:
			err = a.client.WriteSecretKey(ctx, e.Path, e.Key, e.NewValue)
		case diff.OpRemove:
			err = a.client.DeleteSecretKey(ctx, e.Path, e.Key)
		default:
			err = fmt.Errorf("unknown op: %v", e.Op)
		}

		r.Success = err == nil
		r.Err = err
		results = append(results, r)
	}

	return results
}
