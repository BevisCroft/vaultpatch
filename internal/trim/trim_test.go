package trim_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/vaultpatch/internal/trim"
	"github.com/yourusername/vaultpatch/internal/vault"
)

func newMockVault(t *testing.T, secrets map[string]map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/v1/secret/data/")
		switch r.Method {
		case http.MethodGet:
			data, ok := secrets[key]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": data}})
		case http.MethodPost, http.MethodPut:
			var body struct {
				Data map[string]string `json:"data"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			secrets[key] = body.Data
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

func TestApply_DryRun_NoWrites(t *testing.T) {
	secrets := map[string]map[string]string{
		"myapp/config": {"DB_PASSWORD": "s3cr3t", "APP_ENV": "prod"},
	}
	srv := newMockVault(t, secrets)
	defer srv.Close()

	client := newTestClient(t, srv)
	tr := trim.New(client, "secret", true)

	res := tr.Apply(context.Background(), "myapp/config", []string{"DB_*"})

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if len(res.Removed) != 1 || res.Removed[0] != "DB_PASSWORD" {
		t.Errorf("expected [DB_PASSWORD] removed, got %v", res.Removed)
	}
	// dry-run: original data must be unchanged
	if secrets["myapp/config"]["DB_PASSWORD"] != "s3cr3t" {
		t.Error("dry-run should not mutate stored secrets")
	}
}

func TestApply_LiveRemovesKeys(t *testing.T) {
	secrets := map[string]map[string]string{
		"myapp/config": {"DB_PASSWORD": "s3cr3t", "APP_ENV": "prod", "LEGACY_KEY": "old"},
	}
	srv := newMockVault(t, secrets)
	defer srv.Close()

	client := newTestClient(t, srv)
	tr := trim.New(client, "secret", false)

	res := tr.Apply(context.Background(), "myapp/config", []string{"legacy_*"})

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if len(res.Removed) != 1 {
		t.Errorf("expected 1 removed key, got %d", len(res.Removed))
	}
	if _, exists := secrets["myapp/config"]["LEGACY_KEY"]; exists {
		t.Error("LEGACY_KEY should have been removed")
	}
}

func TestApply_NoMatch(t *testing.T) {
	secrets := map[string]map[string]string{
		"myapp/config": {"APP_ENV": "prod"},
	}
	srv := newMockVault(t, secrets)
	defer srv.Close()

	client := newTestClient(t, srv)
	tr := trim.New(client, "secret", false)

	res := tr.Apply(context.Background(), "myapp/config", []string{"DB_*"})

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if len(res.Removed) != 0 {
		t.Errorf("expected no removed keys, got %v", res.Removed)
	}
}

func TestFprint_Summary(t *testing.T) {
	results := []trim.Result{
		{Path: "app/prod", Removed: []string{"OLD_KEY"}, DryRun: true},
		{Path: "app/staging", Removed: []string{}},
	}
	var buf bytes.Buffer
	trim.Fprint(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "would trim") {
		t.Errorf("expected 'would trim' in output, got: %s", out)
	}
	if !strings.Contains(out, "no matching keys") {
		t.Errorf("expected 'no matching keys' in output, got: %s", out)
	}
}
