// Package archive provides functionality for archiving (soft-deleting)
// Vault secrets by moving them to a designated archive path with metadata.
package archive

import (
	"context"
	"fmt"
	"time"

	"github.com/your-org/vaultpatch/internal/vault"
)

// Result holds the outcome of a single archive operation.
type Result struct {
	Path    string
	Archive string
	DryRun  bool
	Err     error
}

// Archiver moves secrets to an archive prefix instead of deleting them.
type Archiver struct {
	client      *vault.Client
	archiveRoot string
	dryRun      bool
}

// New creates an Archiver that stores archived secrets under archiveRoot.
func New(client *vault.Client, archiveRoot string, dryRun bool) *Archiver {
	return &Archiver{
		client:      client,
		archiveRoot: archiveRoot,
		dryRun:      dryRun,
	}
}

// Apply archives each of the given paths by copying the secret data to
// <archiveRoot>/<timestamp>/<original-path> and then deleting the source.
func (a *Archiver) Apply(ctx context.Context, paths []string) []Result {
	results := make([]Result, 0, len(paths))
	ts := time.Now().UTC().Format("20060102T150405Z")

	for _, p := range paths {
		ar := fmt.Sprintf("%s/%s/%s", a.archiveRoot, ts, p)
		result := Result{Path: p, Archive: ar, DryRun: a.dryRun}

		if a.dryRun {
			results = append(results, result)
			continue
		}

		data, err := a.client.ReadSecret(ctx, p)
		if err != nil {
			result.Err = fmt.Errorf("read %s: %w", p, err)
			results = append(results, result)
			continue
		}

		if err := a.client.WriteSecret(ctx, ar, data); err != nil {
			result.Err = fmt.Errorf("write archive %s: %w", ar, err)
			results = append(results, result)
			continue
		}

		if err := a.client.DeleteSecret(ctx, p); err != nil {
			result.Err = fmt.Errorf("delete %s: %w", p, err)
		}

		results = append(results, result)
	}

	return results
}
