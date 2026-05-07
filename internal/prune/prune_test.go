package prune_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/example/vaultpatch/internal/prune"
)

// --- stubs ---

type stubAuditor struct{}

func (s *stubAuditor) Record(_ context.Context, _, _ string, _ bool, _ error) error { return nil }

type stubVault struct {
	paths   []string
	secrets map[string]map[string]string
	deleted []string
}

func (s *stubVault) ListSecrets(_ context.Context, _ string) ([]string, error) {
	return s.paths, nil
}
func (s *stubVault) ReadSecret(_ context.Context, path string) (map[string]string, error) {
	return s.secrets[path], nil
}
func (s *stubVault) DeleteSecret(_ context.Context, path string) error {
	s.deleted = append(s.deleted, path)
	return nil
}

// --- helpers ---

func newMockVault(t *testing.T, paths []string, secrets map[string]map[string]string) *stubVault {
	t.Helper()
	return &stubVault{paths: paths, secrets: secrets}
}

// newMockVaultServer is kept to satisfy the pattern used across other packages
// even though these tests use the stub directly.
func newMockVaultServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]string{}})
	}))
}

// --- tests ---

func TestRun_DryRun_NoDeletes(t *testing.T) {
	old := time.Now().UTC().Add(-48 * time.Hour).Format(time.RFC3339)
	vault := newMockVault(t,
		[]string{"secret/old"},
		map[string]map[string]string{"secret/old": {"updated_at": old, "key": "val"}},
	)

	p := prune.New(vault, &stubAuditor{}, 24*time.Hour, true)
	results, err := p.Run(context.Background(), "secret/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Pruned || !results[0].DryRun {
		t.Errorf("expected DryRun pruned result, got %+v", results[0])
	}
	if len(vault.deleted) != 0 {
		t.Errorf("expected no deletes in dry-run, got %v", vault.deleted)
	}
}

func TestRun_PrunesStaleSecret(t *testing.T) {
	old := time.Now().UTC().Add(-72 * time.Hour).Format(time.RFC3339)
	vault := newMockVault(t,
		[]string{"secret/stale"},
		map[string]map[string]string{"secret/stale": {"updated_at": old}},
	)

	p := prune.New(vault, &stubAuditor{}, 24*time.Hour, false)
	results, err := p.Run(context.Background(), "secret/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || !results[0].Pruned {
		t.Fatalf("expected pruned result, got %+v", results)
	}
	if len(vault.deleted) != 1 || vault.deleted[0] != "secret/stale" {
		t.Errorf("expected secret/stale deleted, got %v", vault.deleted)
	}
}

func TestRun_SkipsFreshSecret(t *testing.T) {
	fresh := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)
	vault := newMockVault(t,
		[]string{"secret/fresh"},
		map[string]map[string]string{"secret/fresh": {"updated_at": fresh}},
	)

	p := prune.New(vault, &stubAuditor{}, 24*time.Hour, false)
	results, err := p.Run(context.Background(), "secret/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || !results[0].Skipped {
		t.Errorf("expected skipped result, got %+v", results)
	}
}

func TestFprint_Summary(t *testing.T) {
	results := []prune.Result{
		{Path: "secret/a", Pruned: true},
		{Path: "secret/b", Skipped: true},
	}
	var buf bytes.Buffer
	prune.Fprint(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "PRUNED") {
		t.Errorf("expected PRUNED in output, got: %s", out)
	}
	if !strings.Contains(out, "SKIP") {
		t.Errorf("expected SKIP in output, got: %s", out)
	}
	if !strings.Contains(out, "Summary:") {
		t.Errorf("expected Summary line in output, got: %s", out)
	}
}
