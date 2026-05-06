// Package verify checks that secrets in Vault match expected values
// defined in a reference map, reporting any mismatches.
package verify

import (
	"context"
	"fmt"

	"github.com/your-org/vaultpatch/internal/vault"
)

// Result holds the outcome of verifying a single secret path.
type Result struct {
	Path    string
	Key     string
	Expected string
	Actual   string
	Match   bool
	Err     error
}

// Verifier compares live Vault secrets against expected values.
type Verifier struct {
	client *vault.Client
	mount  string
}

// New creates a Verifier backed by the given Vault client and KV mount.
func New(client *vault.Client, mount string) *Verifier {
	return &Verifier{client: client, mount: mount}
}

// Check reads each path from Vault and compares it against the expected map.
// expected is keyed by "path/key" and maps to the expected plaintext value.
func (v *Verifier) Check(ctx context.Context, expected map[string]string) ([]Result, error) {
	// Collect unique paths from expected keys.
	paths := map[string]struct{}{}
	for pk := range expected {
		path, _, err := splitPathKey(pk)
		if err != nil {
			return nil, fmt.Errorf("verify: invalid key %q: %w", pk, err)
		}
		paths[path] = struct{}{}
	}

	var results []Result
	for path := range paths {
		secrets, err := v.client.ReadSecret(ctx, v.mount, path)
		if err != nil {
			results = append(results, Result{Path: path, Err: err})
			continue
		}
		for pk, expVal := range expected {
			p, key, _ := splitPathKey(pk)
			if p != path {
				continue
			}
			actual, _ := secrets[key]
			results = append(results, Result{
				Path:     path,
				Key:      key,
				Expected: expVal,
				Actual:   actual,
				Match:    actual == expVal,
			})
		}
	}
	return results, nil
}

// splitPathKey splits a "path/key" string into its two components.
func splitPathKey(pk string) (path, key string, err error) {
	for i := len(pk) - 1; i >= 0; i-- {
		if pk[i] == '/' {
			return pk[:i], pk[i+1:], nil
		}
	}
	return "", "", fmt.Errorf("no '/' separator found")
}
