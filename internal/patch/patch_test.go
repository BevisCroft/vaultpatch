package patch_test

import (
	"context"
	"testing"

	"github.com/example/vaultpatch/internal/diff"
	"github.com/example/vaultpatch/internal/patch"
	"github.com/example/vaultpatch/internal/vault"
)

func newTestApplier(t *testing.T, dryRun bool) (*patch.Applier, func()) {
	t.Helper()
	srv, client := newMockVaultServer(t)
	applier := patch.NewApplier(client, dryRun)
	return applier, srv.Close
}

// newMockVaultServer and newTestClient are shared helpers; reuse from vault tests
// via a thin wrapper so we don't import internal test helpers directly.
func newMockVaultServer(t *testing.T) (interface{ Close() }, *vault.Client) {
	t.Helper()
	// Minimal stub — real integration uses vault_test helpers.
	// For unit tests we just verify dry-run behaviour without a real server.
	return &nopCloser{}, nil
}

type nopCloser struct{}

func (n *nopCloser) Close() {}

func TestApply_DryRun_NoErrors(t *testing.T) {
	entries := []diff.Entry{
		{Path: "secret/app", Key: "DB_PASS", Op: diff.OpAdd, NewValue: "s3cr3t"},
		{Path: "secret/app", Key: "OLD_KEY", Op: diff.OpRemove, OldValue: "gone"},
		{Path: "secret/app", Key: "API", Op: diff.OpUpdate, OldValue: "v1", NewValue: "v2"},
	}

	applier := patch.NewApplier(nil, true) // nil client is safe in dry-run
	results := applier.Apply(context.Background(), entries)

	if len(results) != len(entries) {
		t.Fatalf("expected %d results, got %d", len(entries), len(results))
	}
	for _, r := range results {
		if !r.Success {
			t.Errorf("expected success for %s/%s in dry-run, got err: %v", r.Path, r.Key, r.Err)
		}
	}
}

func TestApply_SkipsOpNone(t *testing.T) {
	entries := []diff.Entry{
		{Path: "secret/app", Key: "UNCHANGED", Op: diff.OpNone},
	}

	applier := patch.NewApplier(nil, true)
	results := applier.Apply(context.Background(), entries)

	if len(results) != 0 {
		t.Fatalf("expected 0 results for OpNone entries, got %d", len(results))
	}
}
