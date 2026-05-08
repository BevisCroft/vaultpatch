// Package resolve provides variable interpolation for Vault secret values,
// allowing references like {{secret/path:key}} to be expanded inline.
package resolve

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// Reader fetches a secret map from a given path.
type Reader interface {
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
}

// Result holds the outcome of resolving a single reference.
type Result struct {
	Ref     string
	Path    string
	Key     string
	Value   string
	Err     error
}

// Resolver expands {{path:key}} placeholders in secret values.
type Resolver struct {
	client Reader
	pattern *regexp.Regexp
}

// New creates a Resolver backed by the given Vault client.
func New(client Reader) *Resolver {
	return &Resolver{
		client:  client,
		pattern: regexp.MustCompile(`\{\{([^}]+):([^}]+)\}\}`),
	}
}

// Apply resolves all {{path:key}} references found in the provided secret map
// and returns a new map with placeholders replaced by their resolved values.
// Unresolvable references are left unchanged and captured in the returned results.
func (r *Resolver) Apply(ctx context.Context, secrets map[string]string) (map[string]string, []Result) {
	out := make(map[string]string, len(secrets))
	var results []Result

	for k, v := range secrets {
		resolved, refs := r.expand(ctx, v)
		out[k] = resolved
		results = append(results, refs...)
	}
	return out, results
}

func (r *Resolver) expand(ctx context.Context, value string) (string, []Result) {
	var results []Result
	result := r.pattern.ReplaceAllStringFunc(value, func(match string) string {
		parts := r.pattern.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		path, key := strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2])
		secrets, err := r.client.ReadSecret(ctx, path)
		if err != nil {
			results = append(results, Result{Ref: match, Path: path, Key: key, Err: err})
			return match
		}
		v, ok := secrets[key]
		if !ok {
			err = fmt.Errorf("key %q not found at path %q", key, path)
			results = append(results, Result{Ref: match, Path: path, Key: key, Err: err})
			return match
		}
		results = append(results, Result{Ref: match, Path: path, Key: key, Value: v})
		return v
	})
	return result, results
}
