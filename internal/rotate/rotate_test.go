package rotate_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/your-org/vaultpatch/internal/audit"
	"github.com/your-org/vaultpatch/internal/rotate"
	"github.com/your-org/vaultpatch/internal/vault"
)

func newMockVault(t *testing.T, secrets map[string]map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path // e.g. /v1/secret/data/app
		switch r.Method {
		case http.MethodGet:
			for k, v := range secrets {
				if "/v1/"+k == path {
					_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": v}})
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		case http.MethodPost, http.MethodPut:
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
	f, err := os.CreateTemp(t.TempDir(), "audit-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return audit.New(f)
}

func TestApply_DryRun_NoWrites(t *testing.T) {
	writes := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": map[string]string{"password": "old"}}})
			return
		}
		writes++
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	client := newTestClient(t, srv)
	rot := rotate.New(client, client, newTestAuditor(t), true)

	gen := func(_, _ string) (string, error) { return "new-value", nil }
	res := rot.Apply(context.Background(), "secret/data/app", gen)

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if writes != 0 {
		t.Errorf("expected 0 writes in dry-run, got %d", writes)
	}
}

func TestApply_LiveRotatesKeys(t *testing.T) {
	written := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": map[string]string{"api_key": "old"}}})
			return
		}
		var body struct {
			Data map[string]string `json:"data"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		for k, v := range body.Data {
			written[k] = v
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	client := newTestClient(t, srv)
	rot := rotate.New(client, client, newTestAuditor(t), false)

	gen := func(_, _ string) (string, error) { return "rotated-value", nil }
	res := rot.Apply(context.Background(), "secret/data/app", gen)

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if len(res.NewKeys) == 0 {
		t.Error("expected at least one rotated key")
	}
}
