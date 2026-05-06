package archive_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/archive"
	"github.com/your-org/vaultpatch/internal/vault"
)

func newMockVault(t *testing.T) *httptest.Server {
	t.Helper()
	store := map[string]map[string]any{
		"secret/data/prod/db": {"password": "s3cr3t"},
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/")
		switch r.Method {
		case http.MethodGet:
			data, ok := store[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": data}})
		case http.MethodPost, http.MethodPut:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			store[path] = body
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			delete(store, path)
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
	srv := newMockVault(t)
	defer srv.Close()

	client := newTestClient(t, srv)
	a := archive.New(client, "secret/data/archive", true)

	results := a.Apply(context.Background(), []string{"secret/data/prod/db"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if !r.DryRun {
		t.Error("expected DryRun=true")
	}
	if r.Err != nil {
		t.Errorf("unexpected error: %v", r.Err)
	}
	if !strings.HasPrefix(r.Archive, "secret/data/archive/") {
		t.Errorf("unexpected archive path: %s", r.Archive)
	}
}

func TestApply_LiveArchivesPaths(t *testing.T) {
	srv := newMockVault(t)
	defer srv.Close()

	client := newTestClient(t, srv)
	a := archive.New(client, "secret/data/archive", false)

	results := a.Apply(context.Background(), []string{"secret/data/prod/db"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Errorf("unexpected error: %v", results[0].Err)
	}
}

func TestFprint_DryRun(t *testing.T) {
	results := []archive.Result{
		{Path: "secret/data/prod/db", Archive: "secret/data/archive/20240101T000000Z/secret/data/prod/db", DryRun: true},
	}
	var buf bytes.Buffer
	archive.Fprint(&buf, results)
	if !strings.Contains(buf.String(), "dry-run") {
		t.Errorf("expected dry-run marker, got: %s", buf.String())
	}
}

func TestFprint_Error(t *testing.T) {
	results := []archive.Result{
		{Path: "secret/data/prod/db", Archive: "secret/data/archive/x", Err: context.DeadlineExceeded},
	}
	var buf bytes.Buffer
	archive.Fprint(&buf, results)
	if !strings.Contains(buf.String(), "error") {
		t.Errorf("expected error marker, got: %s", buf.String())
	}
}
