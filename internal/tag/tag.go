package tag

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Writer can write secret metadata to Vault.
type Writer interface {
	WriteSecretMetadata(ctx context.Context, path string, metadata map[string]string) error
	ReadSecretMetadata(ctx context.Context, path string) (map[string]string, error)
}

// Manager applies and reads tags (metadata labels) on Vault secret paths.
type Manager struct {
	client Writer
	mount  string
	dryRun bool
}

// New creates a Manager.
func New(client Writer, mount string, dryRun bool) *Manager {
	return &Manager{client: client, mount: mount, dryRun: dryRun}
}

// Result holds the outcome of a tag operation on a single path.
type Result struct {
	Path    string
	Tags    map[string]string
	DryRun  bool
	Err     error
}

// Apply sets the provided tags on each path. Existing metadata keys not
// present in tags are preserved (merge semantics).
func (m *Manager) Apply(ctx context.Context, paths []string, tags map[string]string) []Result {
	results := make([]Result, 0, len(paths))
	for _, p := range paths {
		r := Result{Path: p, Tags: tags, DryRun: m.dryRun}
		if !m.dryRun {
			existing, err := m.client.ReadSecretMetadata(ctx, fullPath(m.mount, p))
			if err != nil {
				r.Err = fmt.Errorf("read metadata %s: %w", p, err)
				results = append(results, r)
				continue
			}
			merged := mergeTags(existing, tags)
			if err := m.client.WriteSecretMetadata(ctx, fullPath(m.mount, p), merged); err != nil {
				r.Err = fmt.Errorf("write metadata %s: %w", p, err)
			}
		}
		results = append(results, r)
	}
	return results
}

// List returns tags for the given path.
func (m *Manager) List(ctx context.Context, path string) (map[string]string, error) {
	return m.client.ReadSecretMetadata(ctx, fullPath(m.mount, path))
}

func mergeTags(base, override map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(override))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		out[k] = v
	}
	return out
}

func fullPath(mount, path string) string {
	return strings.TrimRight(mount, "/") + "/" + strings.TrimLeft(path, "/")
}

// SortedKeys returns tag keys in deterministic order.
func SortedKeys(tags map[string]string) []string {
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
