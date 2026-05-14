// Package protect provides path-level write protection for Vault secrets.
// A protected path cannot be modified or deleted without explicitly removing
// the protection metadata first.
package protect

import (
	"context"
	"fmt"
	"time"
)

const metaKey = "vaultpatch/protected"

// VaultClient is the subset of vault operations required by the Protector.
type VaultClient interface {
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
	WriteSecret(ctx context.Context, path string, data map[string]string) error
	DeleteSecret(ctx context.Context, path string) error
}

// Result holds the outcome of a single protect/unprotect operation.
type Result struct {
	Path      string
	Protected bool // true = protect, false = unprotect
	DryRun    bool
	Err       error
}

// Protector applies or removes write-protection metadata on Vault paths.
type Protector struct {
	client VaultClient
	dryRun bool
}

// New creates a Protector backed by the supplied VaultClient.
func New(client VaultClient, dryRun bool) *Protector {
	return &Protector{client: client, dryRun: dryRun}
}

// Protect marks each path as protected by writing a metadata sentinel key.
func (p *Protector) Protect(ctx context.Context, paths []string) []Result {
	return p.apply(ctx, paths, true)
}

// Unprotect removes the protection sentinel from each path.
func (p *Protector) Unprotect(ctx context.Context, paths []string) []Result {
	return p.apply(ctx, paths, false)
}

// IsProtected returns true when the given path carries the protection sentinel.
func (p *Protector) IsProtected(ctx context.Context, path string) (bool, error) {
	data, err := p.client.ReadSecret(ctx, path)
	if err != nil {
		return false, fmt.Errorf("protect: read %s: %w", path, err)
	}
	_, ok := data[metaKey]
	return ok, nil
}

func (p *Protector) apply(ctx context.Context, paths []string, protect bool) []Result {
	results := make([]Result, 0, len(paths))
	for _, path := range paths {
		r := Result{Path: path, Protected: protect, DryRun: p.dryRun}
		if !p.dryRun {
			r.Err = p.write(ctx, path, protect)
		}
		results = append(results, r)
	}
	return results
}

func (p *Protector) write(ctx context.Context, path string, protect bool) error {
	existing, err := p.client.ReadSecret(ctx, path)
	if err != nil {
		return fmt.Errorf("protect: read %s: %w", path, err)
	}
	data := make(map[string]string, len(existing)+1)
	for k, v := range existing {
		data[k] = v
	}
	if protect {
		data[metaKey] = time.Now().UTC().Format(time.RFC3339)
	} else {
		delete(data, metaKey)
	}
	if err := p.client.WriteSecret(ctx, path, data); err != nil {
		return fmt.Errorf("protect: write %s: %w", path, err)
	}
	return nil
}
