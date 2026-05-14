// Package dedupe identifies and removes duplicate secret values across
// multiple Vault paths, helping to reduce redundancy and improve secret hygiene.
package dedupe

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
)

// SecretReader reads secrets from a Vault path.
type SecretReader interface {
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
}

// Result holds information about a duplicate secret value found across paths.
type Result struct {
	Key        string
	Value      string
	Paths      []string
	Fingerprint string
}

// Deduper detects duplicate secret values across Vault paths.
type Deduper struct {
	client SecretReader
	maskValues bool
}

// New creates a new Deduper.
func New(client SecretReader, maskValues bool) *Deduper {
	return &Deduper{client: client, maskValues: maskValues}
}

// Detect scans the given paths and returns any duplicate key/value pairs.
func (d *Deduper) Detect(ctx context.Context, paths []string) ([]Result, error) {
	type entry struct {
		key   string
		value string
	}
	index := make(map[string][]string) // fingerprint -> []"path:key"
	values := make(map[string]entry)   // fingerprint -> entry

	for _, path := range paths {
		secrets, err := d.client.ReadSecret(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("dedupe: read %q: %w", path, err)
		}
		for k, v := range secrets {
			fp := fingerprint(k, v)
			index[fp] = append(index[fp], path)
			values[fp] = entry{key: k, value: v}
		}
	}

	var results []Result
	for fp, paths := range index {
		if len(paths) < 2 {
			continue
		}
		sort.Strings(paths)
		e := values[fp]
		v := e.value
		if d.maskValues {
			v = "***"
		}
		results = append(results, Result{
			Key:         e.key,
			Value:       v,
			Paths:       paths,
			Fingerprint: fp,
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Fingerprint < results[j].Fingerprint
	})
	return results, nil
}

func fingerprint(key, value string) string {
	h := sha256.Sum256([]byte(key + "\x00" + value))
	return fmt.Sprintf("%x", h[:8])
}
