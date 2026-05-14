// Package prefix provides utilities for bulk-renaming Vault secret keys
// by adding or stripping a common prefix across one or more paths.
package prefix

import (
	"context"
	"fmt"
	"strings"
)

// SecretWriter can write a secret to a given path.
type SecretWriter interface {
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
}

// Op describes the kind of prefix operation to perform.
type Op string

const (
	OpAdd    Op = "add"
	OpStrip  Op = "strip"
)

// Result holds the outcome of a prefix operation on a single path.
type Result struct {
	Path    string
	Op      Op
	Prefix  string
	Changed int
	DryRun  bool
	Err     error
}

// Applier applies prefix operations to Vault secrets.
type Applier struct {
	client SecretWriter
	dryRun bool
}

// New returns a new Applier.
func New(client SecretWriter, dryRun bool) *Applier {
	return &Applier{client: client, dryRun: dryRun}
}

// Apply reads the secret at path, renames all keys by adding or stripping
// the given prefix, and writes the result back (unless dry-run).
func (a *Applier) Apply(ctx context.Context, path string, op Op, pfx string) Result {
	res := Result{Path: path, Op: op, Prefix: pfx, DryRun: a.dryRun}

	data, err := a.client.ReadSecret(ctx, path)
	if err != nil {
		res.Err = fmt.Errorf("read %s: %w", path, err)
		return res
	}

	updated := make(map[string]interface{}, len(data))
	for k, v := range data {
		newKey := applyOp(op, pfx, k)
		updated[newKey] = v
		if newKey != k {
			res.Changed++
		}
	}

	if res.Changed == 0 || a.dryRun {
		return res
	}

	if err := a.client.WriteSecret(ctx, path, updated); err != nil {
		res.Err = fmt.Errorf("write %s: %w", path, err)
	}
	return res
}

func applyOp(op Op, pfx, key string) string {
	switch op {
	case OpAdd:
		if !strings.HasPrefix(key, pfx) {
			return pfx + key
		}
	case OpStrip:
		return strings.TrimPrefix(key, pfx)
	}
	return key
}
