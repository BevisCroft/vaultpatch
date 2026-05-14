// Package digest computes and compares cryptographic fingerprints of Vault
// secret paths, making it easy to detect tampering or unexpected mutations.
package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// Result holds the digest output for a single secret path.
type Result struct {
	Path   string
	Digest string
	Match  bool   // populated when compared against a baseline
	Err    error
}

// Digester computes SHA-256 fingerprints over secret key/value maps.
type Digester struct {
	reader SecretReader
}

// SecretReader is the minimal Vault interface required by Digester.
type SecretReader interface {
	ReadSecret(path string) (map[string]string, error)
}

// New returns a Digester backed by the provided reader.
func New(r SecretReader) *Digester {
	return &Digester{reader: r}
}

// Compute returns a deterministic SHA-256 hex digest for the given path.
// Keys are sorted before hashing to ensure stability.
func (d *Digester) Compute(path string) Result {
	secrets, err := d.reader.ReadSecret(path)
	if err != nil {
		return Result{Path: path, Err: err}
	}
	return Result{Path: path, Digest: hashSecrets(secrets)}
}

// Compare computes the current digest for path and checks it against baseline.
func (d *Digester) Compare(path, baseline string) Result {
	r := d.Compute(path)
	if r.Err != nil {
		return r
	}
	r.Match = r.Digest == baseline
	return r
}

// hashSecrets produces a stable SHA-256 hex string from a key/value map.
func hashSecrets(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		fmt.Fprintf(h, "%s=%s\n", k, m[k])
	}
	return strings.ToLower(hex.EncodeToString(h.Sum(nil)))
}
