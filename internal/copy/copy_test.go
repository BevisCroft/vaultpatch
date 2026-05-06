package copy_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	vaultcopy "github.com/yourusername/vaultpatch/internal/copy"
	"github.com/yourusername/vaultpatch/internal/vault"
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
			secrets[path] = body.Data
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

func TestCopy_DryRun_NoWrites(t *testing.T) {
	secrets := map[string]map[string]string{
		"secret/src": {"key": "value"},
	}
	srv := newMockVault(t, secrets)
	defer srv.Close()

	c := vaultcopy.New(newTestClient(t, srv), true)
	result := c.Copy(context.Background(), "secret/src", "secret/dst")

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if !result.DryRun {
		t.Error("expected DryRun=true")
	}
	if _, written := secrets["secret/dst"]; written {
		t.Error("expected no write in dry-run mode")
	}
}

func TestCopy_LiveCopiesData(t *testing.T) {
	secrets := map[string]map[string]string{
		"secret/src": {"foo": "bar"},
	}
	srv := newMockVault(t, secrets)
	defer srv.Close()

	c := vaultcopy.New(newTestClient(t, srv), false)
	result := c.Copy(context.Background(), "secret/src", "secret/dst")

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if got := secrets["secret/dst"]["foo"]; got != "bar" {
		t.Errorf("expected 'bar', got %q", got)
	}
}

func TestCopy_MissingSource_ReturnsError(t *testing.T) {
	srv := newMockVault(t, map[string]map[string]string{})
	defer srv.Close()

	c := vaultcopy.New(newTestClient(t, srv), false)
	result := c.Copy(context.Background(), "secret/missing", "secret/dst")

	if result.Err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestCopyAll_ReturnsAllResults(t *testing.T) {
	secrets := map[string]map[string]string{
		"secret/a": {"x": "1"},
		"secret/b": {"y": "2"},
	}
	srv := newMockVault(t, secrets)
	defer srv.Close()

	c := vaultcopy.New(newTestClient(t, srv), false)
	results := c.CopyAll(context.Background(), [][2]string{
		{"secret/a", "secret/a-copy"},
		{"secret/b", "secret/b-copy"},
	})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s→%s: %v", r.Src, r.Dst, r.Err)
		}
	}
}
