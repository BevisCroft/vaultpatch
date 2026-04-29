// Package drift detects configuration drift between a saved snapshot
// and the live state of secrets in Vault.
package drift

import (
	"context"
	"fmt"

	"github.com/example/vaultpatch/internal/diff"
	"github.com/example/vaultpatch/internal/snapshot"
	"github.com/example/vaultpatch/internal/vault"
)

// Report holds the drift detection result for a single path.
type Report struct {
	Path    string
	Changes []diff.Change
}

// Detector compares a snapshot against live Vault secrets.
type Detector struct {
	client *vault.Client
	mount  string
}

// New creates a new Detector using the provided Vault client and KV mount.
func New(client *vault.Client, mount string) *Detector {
	return &Detector{client: client, mount: mount}
}

// Detect reads every path recorded in snap from Vault and returns drift
// reports for paths whose live values differ from the snapshot.
func (d *Detector) Detect(ctx context.Context, snap *snapshot.Snapshot) ([]Report, error) {
	var reports []Report

	for path, snapData := range snap.Secrets {
		live, err := d.client.ReadSecret(ctx, d.mount, path)
		if err != nil {
			return nil, fmt.Errorf("drift: read %q: %w", path, err)
		}

		changes := diff.Compute(snapData, live)
		if len(changes) > 0 {
			reports = append(reports, Report{Path: path, Changes: changes})
		}
	}

	return reports, nil
}
