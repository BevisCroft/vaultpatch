// Package snapshot provides functionality to capture and restore
// a point-in-time view of Vault KV secrets at a given path.
package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Snapshot holds a captured state of secrets under a Vault path.
type Snapshot struct {
	CapturedAt time.Time            `json:"captured_at"`
	Mount      string               `json:"mount"`
	BasePath   string               `json:"base_path"`
	Secrets    map[string]SecretMap `json:"secrets"`
}

// SecretMap is a key/value map of a single secret's data.
type SecretMap map[string]string

// Lister can list secret paths beneath a prefix.
type Lister interface {
	ListSecrets(ctx context.Context, path string) ([]string, error)
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
}

// Capture walks all secrets under basePath and returns a Snapshot.
func Capture(ctx context.Context, client Lister, mount, basePath string) (*Snapshot, error) {
	snap := &Snapshot{
		CapturedAt: time.Now().UTC(),
		Mount:      mount,
		BasePath:   basePath,
		Secrets:    make(map[string]SecretMap),
	}

	paths, err := client.ListSecrets(ctx, basePath)
	if err != nil {
		return nil, fmt.Errorf("snapshot: list %q: %w", basePath, err)
	}

	for _, p := range paths {
		data, err := client.ReadSecret(ctx, p)
		if err != nil {
			return nil, fmt.Errorf("snapshot: read %q: %w", p, err)
		}
		snap.Secrets[p] = SecretMap(data)
	}

	return snap, nil
}

// Save writes the snapshot as JSON to the given file path.
func Save(snap *Snapshot, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("snapshot: create file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snap); err != nil {
		return fmt.Errorf("snapshot: encode: %w", err)
	}
	return nil
}

// Load reads a snapshot from a JSON file.
func Load(filePath string) (*Snapshot, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("snapshot: open file: %w", err)
	}
	defer f.Close()

	var snap Snapshot
	if err := json.NewDecoder(f).Decode(&snap); err != nil {
		return nil, fmt.Errorf("snapshot: decode: %w", err)
	}
	return &snap, nil
}
