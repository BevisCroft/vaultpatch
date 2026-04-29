// Package template provides functionality for rendering Vault secret values
// into text/template strings, enabling dynamic configuration generation.
package template

import (
	"bytes"
	"fmt"
	"text/template"
)

// Renderer renders Go templates using Vault secret data as the data source.
type Renderer struct {
	secrets map[string]map[string]string
}

// New creates a new Renderer seeded with the provided secrets map.
// The outer key is the Vault path; the inner key is the secret field name.
func New(secrets map[string]map[string]string) *Renderer {
	return &Renderer{secrets: secrets}
}

// Render executes the given template string with the secrets map available
// under the .Secrets key. Returns the rendered output or an error.
func (r *Renderer) Render(tmplStr string) (string, error) {
	funcMap := template.FuncMap{
		"secret": r.lookupSecret,
	}

	t, err := template.New("vaultpatch").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	data := map[string]any{
		"Secrets": r.secrets,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execute error: %w", err)
	}

	return buf.String(), nil
}

// lookupSecret retrieves a secret value by path and key.
// Returns an error string if the path or key is not found.
func (r *Renderer) lookupSecret(path, key string) (string, error) {
	fields, ok := r.secrets[path]
	if !ok {
		return "", fmt.Errorf("secret path %q not found", path)
	}
	val, ok := fields[key]
	if !ok {
		return "", fmt.Errorf("secret key %q not found at path %q", key, path)
	}
	return val, nil
}
