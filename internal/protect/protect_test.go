package protect_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/your-org/vaultpatch/internal/protect"
	"github.com/your-org/vaultpatch/internal/vault"
)

func newMockVault(t *testing.T, secrets map[string]map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/")
		switch r.Method {
		case http.MethodGet:
			data, ok := secrets[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
		case http.MethodPost, http.MethodPut:
			var body struct {
				Data map[string]string `json:"data"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if secrets == nil {
				secrets = map[string]map[string]string{}
			}
			secrets[path] = body.Data
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func newTestClient(t *testing.T, srv *httptest.Server) *vault.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	raw, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("vault client: %v", err)
	}
	raw.SetToken("test-token")
	return vault.NewClient(raw, "secret")
}

func TestProtect_DryRun_NoWrites(t *testing.T) {
	srv := newMockVault(t, map[string]map[string]string{
		"secret/prod/db": {"password": "s3cr3t"},
	})
	defer srv.Close()

	p := protect.New(newTestClient(t, srv), true)
	results := p.Protect(context.Background(), []string{"secret/prod/db"})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if !results[0].DryRun {
		t.Error("expected DryRun=true")
	}
}

func TestProtect_LiveWritesMetadata(t *testing.T) {
	store := map[string]map[string]string{
		"secret/prod/db": {"password": "s3cr3t"},
	}
	srv := newMockVault(t, store)
	defer srv.Close()

	p := protect.New(newTestClient(t, srv), false)
	results := p.Protect(context.Background(), []string{"secret/prod/db"})

	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}

	ok, err := p.IsProtected(context.Background(), "secret/prod/db")
	if err != nil {
		t.Fatalf("IsProtected: %v", err)
	}
	if !ok {
		t.Error("expected path to be protected")
	}
}

func TestUnprotect_RemovesSentinel(t *testing.T) {
	store := map[string]map[string]string{
		"secret/prod/db": {"password": "s3cr3t", "vaultpatch/protected": "2024-01-01T00:00:00Z"},
	}
	srv := newMockVault(t, store)
	defer srv.Close()

	p := protect.New(newTestClient(t, srv), false)
	results := p.Unprotect(context.Background(), []string{"secret/prod/db"})

	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}

	ok, err := p.IsProtected(context.Background(), "secret/prod/db")
	if err != nil {
		t.Fatalf("IsProtected: %v", err)
	}
	if ok {
		t.Error("expected path to be unprotected")
	}
}
