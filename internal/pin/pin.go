// Package pin provides functionality for pinning Vault secret versions,
// preventing them from being overwritten by automated processes until unpinned.
package pin

import (
	"context"
	"fmt"
	"time"

	"github.com/your-org/vaultpatch/internal/vault"
)

// Entry represents a pinned secret path with metadata.
type Entry struct {
	Path      string    `json:"path"`
	PinnedAt  time.Time `json:"pinned_at"`
	PinnedBy  string    `json:"pinned_by"`
	Reason    string    `json:"reason"`
}

// Result holds the outcome of a pin or unpin operation.
type Result struct {
	Path    string
	Op      string // "pin" or "unpin"
	DryRun  bool
	Err     error
}

// Pinner applies and removes pins on secret paths.
type Pinner struct {
	client  *vault.Client
	mount   string
	dryRun  bool
}

// New creates a new Pinner.
func New(client *vault.Client, mount string, dryRun bool) *Pinner {
	return &Pinner{client: client, mount: mount, dryRun: dryRun}
}

// Pin marks the given path as pinned by writing metadata.
func (p *Pinner) Pin(ctx context.Context, path, pinnedBy, reason string) Result {
	r := Result{Path: path, Op: "pin", DryRun: p.dryRun}
	if p.dryRun {
		return r
	}
	metaPath := fmt.Sprintf("%s/metadata/%s", p.mount, path)
	data := map[string]interface{}{
		"custom_metadata": map[string]interface{}{
			"pinned":     "true",
			"pinned_by":  pinnedBy,
			"pinned_at":  time.Now().UTC().Format(time.RFC3339),
			"pin_reason": reason,
		},
	}
	_, err := p.client.Write(ctx, metaPath, data)
	r.Err = err
	return r
}

// Unpin removes the pin metadata from the given path.
func (p *Pinner) Unpin(ctx context.Context, path string) Result {
	r := Result{Path: path, Op: "unpin", DryRun: p.dryRun}
	if p.dryRun {
		return r
	}
	metaPath := fmt.Sprintf("%s/metadata/%s", p.mount, path)
	data := map[string]interface{}{
		"custom_metadata": map[string]interface{}{
			"pinned":     "",
			"pinned_by":  "",
			"pinned_at":  "",
			"pin_reason": "",
		},
	}
	_, err := p.client.Write(ctx, metaPath, data)
	r.Err = err
	return r
}

// IsPinned checks whether a secret path currently has an active pin.
func (p *Pinner) IsPinned(ctx context.Context, path string) (bool, error) {
	metaPath := fmt.Sprintf("%s/metadata/%s", p.mount, path)
	secret, err := p.client.Read(ctx, metaPath)
	if err != nil {
		return false, err
	}
	if secret == nil {
		return false, nil
	}
	cm, ok := secret["custom_metadata"].(map[string]interface{})
	if !ok {
		return false, nil
	}
	return cm["pinned"] == "true", nil
}
