package cleanup_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/vaultpatch/internal/audit"
	"github.com/your-org/vaultpatch/internal/cleanup"
	"github.com/your-org/vaultpatch/internal/vault"
)

func newMockVault(t *testing.T, secrets map[string]map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			path := r.URL.Path
			if r.URL.Query().Get("list") == "true" {
				keys := make([]string, 0, len(secrets))
				for k := range secrets {
					keys = append(keys, k)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"keys": keys}})
				return
			}
			for k, v := range secrets {
				if "/v1/secret/data/"+k == path {
					_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": v}})
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func newTestClient(t *testing.T, srv *httptest.Server) *vault.Client {
	t.Helper()
	c, err := vault.NewClient(srv.URL, "test-token", "secret")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func newTestAuditor(t *testing.T) *audit.Auditor {
	t.Helper()
	var buf bytes.Buffer
	return audit.New(&buf)
}

func TestRun_DryRun_NoDeletes(t *testing.T) {
	oldTime := time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339)
	srv := newMockVault(t, map[string]map[string]string{
		"stale-key": {"value": "old", "updated_at": oldTime},
	})
	defer srv.Close()

	cleaner := cleanup.New(newTestClient(t, srv), newTestAuditor(t), 24*time.Hour, true)
	results, err := cleaner.Run(context.Background(), "prefix")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Deleted {
		t.Error("expected no deletion in dry-run mode")
	}
	if !results[0].DryRun {
		t.Error("expected DryRun flag to be set")
	}
}

func TestRun_SkipsFreshSecrets(t *testing.T) {
	newTime := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
	srv := newMockVault(t, map[string]map[string]string{
		"fresh-key": {"value": "new", "updated_at": newTime},
	})
	defer srv.Close()

	cleaner := cleanup.New(newTestClient(t, srv), newTestAuditor(t), 24*time.Hour, false)
	results, err := cleaner.Run(context.Background(), "prefix")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for fresh secret, got %d", len(results))
	}
}

func TestFprint_DryRun(t *testing.T) {
	var buf bytes.Buffer
	results := []cleanup.Result{
		{Path: "secret/old", Age: 50 * time.Hour, DryRun: true},
	}
	cleanup.Fprint(&buf, results)
	if got := buf.String(); got == "" {
		t.Error("expected non-empty output")
	}
}

func TestFprint_NoResults(t *testing.T) {
	var buf bytes.Buffer
	cleanup.Fprint(&buf, nil)
	if got := buf.String(); got == "" {
		t.Error("expected no-results message")
	}
}
