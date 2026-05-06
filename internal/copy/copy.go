package copy

import (
	"context"
	"fmt"
)

// SecretReader reads a secret from a given path.
type SecretReader interface {
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
}

// SecretWriter writes a secret to a given path.
type SecretWriter interface {
	WriteSecret(ctx context.Context, path string, data map[string]string) error
}

// Client combines read and write capabilities.
type Client interface {
	SecretReader
	SecretWriter
}

// Result holds the outcome of a single copy operation.
type Result struct {
	Src    string
	Dst    string
	DryRun bool
	Err    error
}

// Copier copies secrets from one path to another.
type Copier struct {
	client Client
	dryRun bool
}

// New returns a new Copier.
func New(client Client, dryRun bool) *Copier {
	return &Copier{client: client, dryRun: dryRun}
}

// Copy reads the secret at src and writes it to dst.
// When dryRun is true no write is performed.
func (c *Copier) Copy(ctx context.Context, src, dst string) Result {
	result := Result{Src: src, Dst: dst, DryRun: c.dryRun}

	data, err := c.client.ReadSecret(ctx, src)
	if err != nil {
		result.Err = fmt.Errorf("read %q: %w", src, err)
		return result
	}

	if c.dryRun {
		return result
	}

	if err := c.client.WriteSecret(ctx, dst, data); err != nil {
		result.Err = fmt.Errorf("write %q: %w", dst, err)
	}
	return result
}

// CopyAll copies multiple src→dst pairs, collecting all results.
func (c *Copier) CopyAll(ctx context.Context, pairs [][2]string) []Result {
	results := make([]Result, 0, len(pairs))
	for _, p := range pairs {
		results = append(results, c.Copy(ctx, p[0], p[1]))
	}
	return results
}
