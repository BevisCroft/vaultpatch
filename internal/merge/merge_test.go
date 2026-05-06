package merge_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/merge"
	"github.com/your-org/vaultpatch/internal/vault"
)

func newMockVault(t *testing.T, data map[string]map[string]string) (*httptest.Server, *vault.Client) {
	t.Helper()
	store := make(map[string]map[string]string)
	for k, v := range data {
		store[k] = v
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/")
		switch r.Method {
		case http.MethodGet:
			secrets, ok := store[path]
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": secrets}})
		case http.MethodPost, http.MethodPut:
			var body struct {
				Data map[string]string `json:"data"`
			}
			json.NewDecoder(r.Body).Decode(&body)
			store[path] = body.Data
			w.WriteHeader(http.StatusNoContent)
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client, err := vault.NewClient(srv.URL, "test-token", "secret")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return srv, client
}

func TestMerge_DryRun_NoWrites(t *testing.T) {
	_, client := newMockVault(t, map[string]map[string]string{
		"secret/data/src": {"a": "1", "b": "2"},
		"secret/data/dst": {"c": "3"},
	})
	m := merge.New(client, merge.StrategyTheirs, true)
	res := m.Apply("secret/data/src", "secret/data/dst")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.DryRun {
		t.Error("expected DryRun=true")
	}
	if len(res.Merged) != 3 {
		t.Errorf("want 3 merged keys, got %d", len(res.Merged))
	}
}

func TestMerge_StrategyTheirs_OverwritesConflict(t *testing.T) {
	_, client := newMockVault(t, map[string]map[string]string{
		"secret/data/src": {"key": "new"},
		"secret/data/dst": {"key": "old"},
	})
	m := merge.New(client, merge.StrategyTheirs, false)
	res := m.Apply("secret/data/src", "secret/data/dst")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Merged["key"] != "new" {
		t.Errorf("want 'new', got %q", res.Merged["key"])
	}
	if len(res.Conflicts) != 1 || res.Conflicts[0] != "key" {
		t.Errorf("expected conflict on 'key', got %v", res.Conflicts)
	}
}

func TestMerge_StrategyOurs_KeepsDestination(t *testing.T) {
	_, client := newMockVault(t, map[string]map[string]string{
		"secret/data/src": {"key": "new"},
		"secret/data/dst": {"key": "old"},
	})
	m := merge.New(client, merge.StrategyOurs, false)
	res := m.Apply("secret/data/src", "secret/data/dst")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Merged["key"] != "old" {
		t.Errorf("want 'old', got %q", res.Merged["key"])
	}
}

func TestMerge_StrategyError_ReturnsErrorOnConflict(t *testing.T) {
	_, client := newMockVault(t, map[string]map[string]string{
		"secret/data/src": {"key": "new"},
		"secret/data/dst": {"key": "old"},
	})
	m := merge.New(client, merge.StrategyError, false)
	res := m.Apply("secret/data/src", "secret/data/dst")
	if res.Err == nil {
		t.Fatal("expected error for conflict, got nil")
	}
}
