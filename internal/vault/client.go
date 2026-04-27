package vault

import (
	"context"
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the Vault API client with helper methods.
type Client struct {
	api  *vaultapi.Client
	Mount string
}

// Config holds configuration for connecting to Vault.
type Config struct {
	Address string
	Token   string
	Mount   string
}

// NewClient creates a new Vault client from the given config.
func NewClient(cfg Config) (*Client, error) {
	apiCfg := vaultapi.DefaultConfig()
	apiCfg.Address = cfg.Address

	api, err := vaultapi.NewClient(apiCfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault api client: %w", err)
	}

	api.SetToken(cfg.Token)

	return &Client{
		api:   api,
		Mount: cfg.Mount,
	}, nil
}

// ReadSecret reads a KV v2 secret at the given path.
func (c *Client) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	secret, err := c.api.KVv2(c.Mount).Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading secret at %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return map[string]interface{}{}, nil
	}
	return secret.Data, nil
}

// WriteSecret writes key-value pairs to the given path.
func (c *Client) WriteSecret(ctx context.Context, path string, data map[string]interface{}) error {
	_, err := c.api.KVv2(c.Mount).Put(ctx, path, data)
	if err != nil {
		return fmt.Errorf("writing secret at %q: %w", path, err)
	}
	return nil
}

// ListSecrets lists all secret keys under the given path prefix.
func (c *Client) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	secret, err := c.api.KVv2(c.Mount).List(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("listing secrets at %q: %w", prefix, err)
	}
	if secret == nil || secret.Data == nil {
		return []string{}, nil
	}
	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return []string{}, nil
	}
	result := make([]string, 0, len(keys))
	for _, k := range keys {
		if s, ok := k.(string); ok {
			result = append(result, s)
		}
	}
	return result, nil
}
