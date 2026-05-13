package status_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/yourusername/vaultpatch/internal/status"
	"github.com/yourusername/vaultpatch/internal/vault"
)

func newMockVault(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
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
	return vault.NewClientFromRaw(raw)
}

func TestCheck_AllOK(t *testing.T) {
	srv := newMockVault(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"data":{"key":"value"}}}`))
	})

	checker := status.New(newTestClient(t, srv))
	results := checker.Check(context.Background(), []string{"secret/a", "secret/b"})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.State != status.StateOK {
			t.Errorf("path %s: expected StateOK, got %v", r.Path, r.State)
		}
	}
}

func TestCheck_MissingPath(t *testing.T) {
	srv := newMockVault(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	checker := status.New(newTestClient(t, srv))
	results := checker.Check(context.Background(), []string{"secret/missing"})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].State != status.StateMissing {
		t.Errorf("expected StateMissing, got %v", results[0].State)
	}
}

func TestCheck_ErrorPath(t *testing.T) {
	srv := newMockVault(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	checker := status.New(newTestClient(t, srv))
	results := checker.Check(context.Background(), []string{"secret/broken"})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].State != status.StateError {
		t.Errorf("expected StateError, got %v", results[0].State)
	}
	if results[0].Err == nil {
		t.Error("expected non-nil Err for StateError result")
	}
}

func TestCheck_SortedOutput(t *testing.T) {
	srv := newMockVault(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"data":{"k":"v"}}}`))
	})

	checker := status.New(newTestClient(t, srv))
	paths := []string{"secret/z", "secret/a", "secret/m"}
	results := checker.Check(context.Background(), paths)

	expected := []string{"secret/a", "secret/m", "secret/z"}
	for i, r := range results {
		if r.Path != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], r.Path)
		}
	}
}

func TestCheck_EmptyPaths(t *testing.T) {
	checker := status.New(stubReader{})
	results := checker.Check(context.Background(), nil)
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

type stubReader struct{}

func (stubReader) ReadSecret(_ context.Context, _ string) (map[string]string, error) {
	return nil, errors.New("secret not found")
}
