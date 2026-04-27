package vault_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/vault"
)

func newMockVaultServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// KV v2 read
	mux.HandleFunc("/v1/secret/data/myapp/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"data": map[string]interface{}{
					"DB_HOST": "localhost",
					"DB_PORT": "5432",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	})

	// KV v2 list
	mux.HandleFunc("/v1/secret/metadata/myapp", func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "list=true") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"keys": []interface{}{"config", "creds"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	})

	return httptest.NewServer(mux)
}

func newTestClient(t *testing.T, addr string) *vault.Client {
	t.Helper()
	c, err := vault.NewClient(vault.Config{
		Address: addr,
		Token:   "test-token",
		Mount:   "secret",
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestNewClient_InvalidAddress(t *testing.T) {
	_, err := vault.NewClient(vault.Config{
		Address: "://bad-address",
		Token:   "tok",
		Mount:   "secret",
	})
	if err == nil {
		t.Fatal("expected error for invalid address, got nil")
	}
}

func TestReadSecret(t *testing.T) {
	srv := newMockVaultServer(t)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	data, err := client.ReadSecret(context.Background(), "myapp/config")
	if err != nil {
		t.Fatalf("ReadSecret: %v", err)
	}
	if data["DB_HOST"] != "localhost" {
		t.Errorf("expected DB_HOST=localhost, got %v", data["DB_HOST"])
	}
}

func TestListSecrets(t *testing.T) {
	srv := newMockVaultServer(t)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	keys, err := client.ListSecrets(context.Background(), "myapp")
	if err != nil {
		t.Fatalf("ListSecrets: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}
