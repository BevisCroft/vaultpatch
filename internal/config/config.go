// Package config handles loading and validating vaultpatch configuration
// from environment variables and optional config files.
package config

import (
	"errors"
	"os"
	"strings"
)

// Config holds all runtime configuration for vaultpatch.
type Config struct {
	// VaultAddr is the address of the Vault server (e.g. https://vault.example.com).
	VaultAddr string

	// VaultToken is the authentication token used to communicate with Vault.
	VaultToken string

	// MountPath is the KV secrets engine mount path (default: "secret").
	MountPath string

	// DryRun controls whether patch operations are simulated without writing.
	DryRun bool

	// MaskSecrets controls whether secret values are masked in diff output.
	MaskSecrets bool
}

// FromEnv loads a Config from well-known environment variables.
// VAULT_ADDR and VAULT_TOKEN are required; others are optional.
func FromEnv() (*Config, error) {
	addr := strings.TrimSpace(os.Getenv("VAULT_ADDR"))
	token := strings.TrimSpace(os.Getenv("VAULT_TOKEN"))

	if addr == "" {
		return nil, errors.New("config: VAULT_ADDR environment variable is required")
	}
	if token == "" {
		return nil, errors.New("config: VAULT_TOKEN environment variable is required")
	}

	mount := os.Getenv("VAULTPATCH_MOUNT")
	if mount == "" {
		mount = "secret"
	}

	return &Config{
		VaultAddr:   addr,
		VaultToken:  token,
		MountPath:   mount,
		DryRun:      envBool("VAULTPATCH_DRY_RUN"),
		MaskSecrets: envBool("VAULTPATCH_MASK_SECRETS"),
	}, nil
}

// Validate returns an error if the Config contains invalid values.
func (c *Config) Validate() error {
	if c.VaultAddr == "" {
		return errors.New("config: VaultAddr must not be empty")
	}
	if c.VaultToken == "" {
		return errors.New("config: VaultToken must not be empty")
	}
	if c.MountPath == "" {
		return errors.New("config: MountPath must not be empty")
	}
	return nil
}

// envBool returns true when the named environment variable is set to
// a truthy value ("1", "true", "yes" — case-insensitive).
func envBool(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes"
}
