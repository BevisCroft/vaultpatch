package clone

import (
	"context"
	"fmt"

	"github.com/your-org/vaultpatch/internal/audit"
)

// Cloner copies secrets from one Vault path to another.
type Cloner struct {
	reader  SecretReader
	writer  SecretWriter
	auditor *audit.Auditor
	dryRun  bool
}

// SecretReader reads a secret from Vault.
type SecretReader interface {
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
}

// SecretWriter writes a secret to Vault.
type SecretWriter interface {
	WriteSecret(ctx context.Context, path string, data map[string]string) error
}

// Result holds the outcome of a single clone operation.
type Result struct {
	Src  string
	Dst  string
	Keys int
	Err  error
}

// New creates a new Cloner.
func New(r SecretReader, w SecretWriter, a *audit.Auditor, dryRun bool) *Cloner {
	return &Cloner{reader: r, writer: w, auditor: a, dryRun: dryRun}
}

// Clone copies all key/value pairs from src to dst.
func (c *Cloner) Clone(ctx context.Context, src, dst string) Result {
	res := Result{Src: src, Dst: dst}

	secrets, err := c.reader.ReadSecret(ctx, src)
	if err != nil {
		res.Err = fmt.Errorf("read %s: %w", src, err)
		_ = c.auditor.Record(ctx, "clone", src, dst, c.dryRun, res.Err)
		return res
	}

	res.Keys = len(secrets)

	if !c.dryRun {
		if err := c.writer.WriteSecret(ctx, dst, secrets); err != nil {
			res.Err = fmt.Errorf("write %s: %w", dst, err)
			_ = c.auditor.Record(ctx, "clone", src, dst, c.dryRun, res.Err)
			return res
		}
	}

	_ = c.auditor.Record(ctx, "clone", src, dst, c.dryRun, nil)
	return res
}
