package config_test

import (
	"testing"

	"github.com/your-org/vaultpatch/internal/config"
)

func TestFromEnv_MissingAddr(t *testing.T) {
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "s.test")

	_, err := config.FromEnv()
	if err == nil {
		t.Fatal("expected error when VAULT_ADDR is missing, got nil")
	}
}

func TestFromEnv_MissingToken(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
	t.Setenv("VAULT_TOKEN", "")

	_, err := config.FromEnv()
	if err == nil {
		t.Fatal("expected error when VAULT_TOKEN is missing, got nil")
	}
}

func TestFromEnv_DefaultMount(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
	t.Setenv("VAULT_TOKEN", "s.test")
	t.Setenv("VAULTPATCH_MOUNT", "")

	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MountPath != "secret" {
		t.Errorf("expected default mount 'secret', got %q", cfg.MountPath)
	}
}

func TestFromEnv_CustomMount(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
	t.Setenv("VAULT_TOKEN", "s.test")
	t.Setenv("VAULTPATCH_MOUNT", "kv")

	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MountPath != "kv" {
		t.Errorf("expected mount 'kv', got %q", cfg.MountPath)
	}
}

func TestFromEnv_BoolFlags(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
	t.Setenv("VAULT_TOKEN", "s.test")
	t.Setenv("VAULTPATCH_DRY_RUN", "true")
	t.Setenv("VAULTPATCH_MASK_SECRETS", "1")

	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.DryRun {
		t.Error("expected DryRun to be true")
	}
	if !cfg.MaskSecrets {
		t.Error("expected MaskSecrets to be true")
	}
}

func TestValidate_EmptyToken(t *testing.T) {
	cfg := &config.Config{
		VaultAddr: "http://127.0.0.1:8200",
		VaultToken: "",
		MountPath:  "secret",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for empty token")
	}
}
