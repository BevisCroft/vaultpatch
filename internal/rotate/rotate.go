package rotate

import (
	"context"
	"fmt"
	"time"

	"github.com/your-org/vaultpatch/internal/audit"
)

// Result holds the outcome of a single secret rotation.
type Result struct {
	Path    string
	OldKeys []string
	NewKeys []string
	DryRun  bool
	Err     error
}

// Generator is a function that produces a new secret value for a given key.
type Generator func(path, key string) (string, error)

// Rotator reads secrets from Vault, replaces values via a Generator, and
// writes the updated data back.
type Rotator struct {
	reader  secretReader
	writer  secretWriter
	auditor *audit.Auditor
	dryRun  bool
}

type secretReader interface {
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
}

type secretWriter interface {
	WriteSecret(ctx context.Context, path string, data map[string]string) error
}

// New creates a Rotator.
func New(r secretReader, w secretWriter, a *audit.Auditor, dryRun bool) *Rotator {
	return &Rotator{reader: r, writer: w, auditor: a, dryRun: dryRun}
}

// Apply rotates the secrets at path using gen to produce replacement values.
func (r *Rotator) Apply(ctx context.Context, path string, gen Generator) Result {
	res := Result{Path: path, DryRun: r.dryRun}

	current, err := r.reader.ReadSecret(ctx, path)
	if err != nil {
		res.Err = fmt.Errorf("read %s: %w", path, err)
		_ = r.auditor.Record(ctx, audit.Entry{
			Path:      path,
			Op:        "rotate",
			DryRun:    r.dryRun,
			Timestamp: time.Now().UTC(),
			Message:   res.Err.Error(),
		})
		return res
	}

	updated := make(map[string]string, len(current))
	for k := range current {
		res.OldKeys = append(res.OldKeys, k)
		newVal, genErr := gen(path, k)
		if genErr != nil {
			res.Err = fmt.Errorf("generate %s/%s: %w", path, k, genErr)
			return res
		}
		updated[k] = newVal
		res.NewKeys = append(res.NewKeys, k)
	}

	if !r.dryRun {
		if err = r.writer.WriteSecret(ctx, path, updated); err != nil {
			res.Err = fmt.Errorf("write %s: %w", path, err)
		}
	}

	_ = r.auditor.Record(ctx, audit.Entry{
		Path:      path,
		Op:        "rotate",
		DryRun:    r.dryRun,
		Timestamp: time.Now().UTC(),
	})
	return res
}
