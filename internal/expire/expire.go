// Package expire provides secret expiry tracking and detection for Vault paths.
package expire

import (
	"context"
	"fmt"
	"time"

	"github.com/youorg/vaultpatch/internal/vault"
)

// Entry represents an expiry record for a single Vault path.
type Entry struct {
	Path      string    `json:"path"`
	ExpiresAt time.Time `json:"expires_at"`
	Note      string    `json:"note,omitempty"`
}

// Result holds the expiry status for a single path.
type Result struct {
	Entry
	Expired bool
	DaysLeft int
}

// Checker checks secrets against their expiry metadata.
type Checker struct {
	client *vault.Client
	mount  string
	now    func() time.Time
}

// New creates a new Checker.
func New(client *vault.Client, mount string) *Checker {
	return &Checker{
		client: client,
		mount:  mount,
		now:    time.Now,
	}
}

// Check reads the expiry metadata key "_expires_at" from each path and
// returns a Result for every path provided.
func (c *Checker) Check(ctx context.Context, paths []string) ([]Result, error) {
	results := make([]Result, 0, len(paths))
	for _, p := range paths {
		fullPath := fmt.Sprintf("%s/%s", c.mount, p)
		secrets, err := c.client.ReadSecret(ctx, fullPath)
		if err != nil {
			return nil, fmt.Errorf("expire: read %q: %w", fullPath, err)
		}
		raw, ok := secrets["_expires_at"]
		if !ok {
			continue
		}
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return nil, fmt.Errorf("expire: parse timestamp for %q: %w", p, err)
		}
		note, _ := secrets["_expire_note"]
		now := c.now()
		daysLeft := int(t.Sub(now).Hours() / 24)
		results = append(results, Result{
			Entry:    Entry{Path: p, ExpiresAt: t, Note: note},
			Expired:  now.After(t),
			DaysLeft: daysLeft,
		})
	}
	return results, nil
}
