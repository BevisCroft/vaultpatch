package search

import (
	"context"
	"strings"

	"github.com/your-org/vaultpatch/internal/vault"
)

// Result holds a single match from a secret search.
type Result struct {
	Path  string
	Key   string
	Value string
}

// Searcher queries Vault secrets for keys or values matching a pattern.
type Searcher struct {
	client *vault.Client
	mount  string
}

// New creates a Searcher backed by the given Vault client.
func New(client *vault.Client, mount string) *Searcher {
	return &Searcher{client: client, mount: mount}
}

// Find walks all secrets under mount and returns entries whose key or value
// contains the query string (case-insensitive). Dry-run is not applicable here;
// search is always read-only.
func (s *Searcher) Find(ctx context.Context, query string, searchValues bool) ([]Result, error) {
	paths, err := s.client.ListSecrets(ctx, s.mount+"/")
	if err != nil {
		return nil, err
	}

	var results []Result
	lower := strings.ToLower(query)

	for _, p := range paths {
		fullPath := s.mount + "/" + p
		secrets, err := s.client.ReadSecret(ctx, fullPath)
		if err != nil {
			continue
		}
		for k, v := range secrets {
			keyMatch := strings.Contains(strings.ToLower(k), lower)
			valMatch := searchValues && strings.Contains(strings.ToLower(v), lower)
			if keyMatch || valMatch {
				results = append(results, Result{
					Path:  fullPath,
					Key:   k,
					Value: v,
				})
			}
		}
	}
	return results, nil
}
