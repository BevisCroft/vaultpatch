// Package rename provides functionality to rename (move) secret paths
// within a Vault mount, optionally performing a dry run.
package rename

import (
	"context"
	"fmt"
	"io"

	"github.com/your-org/vaultpatch/internal/audit"
)

// VaultClient abstracts the Vault operations needed for renaming.
type VaultClient interface {
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
	DeleteSecret(ctx context.Context, path string) error
}

// Result holds the outcome of a single rename operation.
type Result struct {
	Src     string
	Dst     string
	DryRun  bool
	Err     error
}

// Renamer moves secrets from one path to another.
type Renamer struct {
	client VaultClient
	auditor *audit.Auditor
	dryRun  bool
}

// New creates a new Renamer.
func New(client VaultClient, auditor *audit.Auditor, dryRun bool) *Renamer {
	return &Renamer{client: client, auditor: auditor, dryRun: dryRun}
}

// Rename reads the secret at src, writes it to dst, then deletes src.
// If dryRun is true, no mutations are performed.
func (r *Renamer) Rename(ctx context.Context, src, dst string) Result {
	res := Result{Src: src, Dst: dst, DryRun: r.dryRun}

	data, err := r.client.ReadSecret(ctx, src)
	if err != nil {
		res.Err = fmt.Errorf("read %q: %w", src, err)
		r.record(src, dst, res.Err)
		return res
	}

	if !r.dryRun {
		if err := r.client.WriteSecret(ctx, dst, data); err != nil {
			res.Err = fmt.Errorf("write %q: %w", dst, err)
			r.record(src, dst, res.Err)
			return res
		}
		if err := r.client.DeleteSecret(ctx, src); err != nil {
			res.Err = fmt.Errorf("delete %q: %w", src, err)
			r.record(src, dst, res.Err)
			return res
		}
	}

	r.record(src, dst, nil)
	return res
}

// Fprint writes a human-readable summary of the result to w.
func Fprint(w io.Writer, res Result) {
	prefix := ""
	if res.DryRun {
		prefix = "[dry-run] "
	}
	if res.Err != nil {
		fmt.Fprintf(w, "%sERROR rename %q -> %q: %v\n", prefix, res.Src, res.Dst, res.Err)
		return
	}
	fmt.Fprintf(w, "%srenamed %q -> %q\n", prefix, res.Src, res.Dst)
}

func (r *Renamer) record(src, dst string, err error) {
	if r.auditor == nil {
		return
	}
	msg := fmt.Sprintf("rename %s -> %s", src, dst)
	r.auditor.Record(audit.Entry{
		Operation: "rename",
		Path:      src,
		Message:   msg,
		DryRun:    r.dryRun,
		Success:   err == nil,
	})
}
