package drift_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/vaultpatch/internal/drift"
	"github.com/example/vaultpatch/internal/snapshot"
	"github.com/example/vaultpatch/internal/vault"
)

func newMockVault(t *testing.T, secrets map[string]map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for path, data := range secrets {
			if r.URL.Path == "/v1/secret/data/"+path {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{"data": data},
				})
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func newTestClient(t *testing.T, addr string) *vault.Client {
	t.Helper()
	c, err := vault.NewClient(addr, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestDetect_NoDrift(t *testing.T) {
	data := map[string]string{"key": "value"}
	srv := newMockVault(t, map[string]map[string]string{"app/config": data})
	defer srv.Close()

	snap := &snapshot.Snapshot{Secrets: map[string]map[string]string{"app/config": data}}
	d := drift.New(newTestClient(t, srv.URL), "secret")

	reports, err := d.Detect(context.Background(), snap)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(reports) != 0 {
		t.Fatalf("expected no drift, got %d reports", len(reports))
	}
}

func TestDetect_DriftDetected(t *testing.T) {
	srv := newMockVault(t, map[string]map[string]string{
		"app/config": {"key": "new-value"},
	})
	defer srv.Close()

	snap := &snapshot.Snapshot{Secrets: map[string]map[string]string{
		"app/config": {"key": "old-value"},
	}}
	d := drift.New(newTestClient(t, srv.URL), "secret")

	reports, err := d.Detect(context.Background(), snap)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 drift report, got %d", len(reports))
	}
	if reports[0].Path != "app/config" {
		t.Errorf("unexpected path %q", reports[0].Path)
	}
	if len(reports[0].Changes) == 0 {
		t.Error("expected changes in report")
	}
}
