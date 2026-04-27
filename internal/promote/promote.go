// Package promote handles copying secrets from one Vault environment to another.
package promote

import (
	"context"
	"fmt"

	"github.com/your-org/vaultpatch/internal/diff"
	"github.com/your-org/vaultpatch/internal/vault"
)

// Options controls the behaviour of a promotion run.
type Options struct {
	// DryRun prevents any writes when true.
	DryRun bool
	// Overwrite controls whether existing keys in the destination are replaced.
	Overwrite bool
}

// Result captures the outcome of a single promoted secret path.
type Result struct {
	Path    string
	Changes []diff.Change
	Err     error
}

// Promoter copies secrets from a source Vault client to a destination Vault client.
type Promoter struct {
	src  *vault.Client
	dst  *vault.Client
	opts Options
}

// New creates a Promoter with the given source, destination and options.
func New(src, dst *vault.Client, opts Options) *Promoter {
	return &Promoter{src: src, dst: dst, opts: opts}
}

// Promote copies all secrets found under path from src to dst.
// It returns one Result per secret path visited.
func (p *Promoter) Promote(ctx context.Context, path string) ([]Result, error) {
	paths, err := p.src.ListSecrets(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("promote: list %q: %w", path, err)
	}

	var results []Result
	for _, sp := range paths {
		res := p.promoteOne(ctx, sp)
		results = append(results, res)
	}
	return results, nil
}

func (p *Promoter) promoteOne(ctx context.Context, path string) Result {
	res := Result{Path: path}

	srcData, err := p.src.ReadSecret(ctx, path)
	if err != nil {
		res.Err = fmt.Errorf("read src: %w", err)
		return res
	}

	dstData, _ := p.dst.ReadSecret(ctx, path) // missing dst is fine — treat as empty

	res.Changes = diff.Compute(dstData, srcData)

	if p.opts.DryRun {
		return res
	}

	merged := mergeMaps(dstData, srcData, p.opts.Overwrite)
	if err := p.dst.WriteSecret(ctx, path, merged); err != nil {
		res.Err = fmt.Errorf("write dst: %w", err)
	}
	return res
}

// mergeMaps combines src into dst. When overwrite is true, src values win.
func mergeMaps(dst, src map[string]string, overwrite bool) map[string]string {
	out := make(map[string]string, len(dst))
	for k, v := range dst {
		out[k] = v
	}
	for k, v := range src {
		if _, exists := out[k]; !exists || overwrite {
			out[k] = v
		}
	}
	return out
}
