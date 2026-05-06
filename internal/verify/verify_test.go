package verify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/vaultpatch/internal/vault"
	"github.com/your-org/vaultpatch/internal/verify"
)

func newMockVault(t *testing.T, secrets map[string]map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for path, data := range secrets {
			if r.URL.Path == "/v1/secret/data/"+path {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{"data": data},
				})
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func newTestClient(t *testing.T, srv *httptest.Server) *vault.Client {
	t.Helper()
	c, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestCheck_AllMatch(t *testing.T) {
	srv := newMockVault(t, map[string]map[string]string{
		"myapp/config": {"host": "localhost", "port": "5432"},
	})
	defer srv.Close()

	v := verify.New(newTestClient(t, srv), "secret")
	results, err := v.Check(context.Background(), map[string]string{
		"myapp/config/host": "localhost",
		"myapp/config/port": "5432",
	})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	for _, r := range results {
		if !r.Match {
			t.Errorf("expected match for %s/%s", r.Path, r.Key)
		}
	}
}

func TestCheck_Mismatch(t *testing.T) {
	srv := newMockVault(t, map[string]map[string]string{
		"myapp/config": {"host": "remotehost"},
	})
	defer srv.Close()

	v := verify.New(newTestClient(t, srv), "secret")
	results, err := v.Check(context.Background(), map[string]string{
		"myapp/config/host": "localhost",
	})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Match {
		t.Error("expected mismatch")
	}
}

func TestCheck_InvalidKey(t *testing.T) {
	srv := newMockVault(t, nil)
	defer srv.Close()

	v := verify.New(newTestClient(t, srv), "secret")
	_, err := v.Check(context.Background(), map[string]string{
		"noslash": "value",
	})
	if err == nil {
		t.Fatal("expected error for key without '/'")
	}
}
