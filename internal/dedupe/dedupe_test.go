package dedupe_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/seatgeek/vaultpatch/internal/dedupe"
)

type stubReader struct {
	data map[string]map[string]string
}

func (s *stubReader) ReadSecret(_ context.Context, path string) (map[string]string, error) {
	return s.data[path], nil
}

func newMockVault(data map[string]map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for path, secrets := range data {
			if r.URL.Path == "/v1/"+path {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{"data": secrets},
				})
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func newTestClient(t *testing.T, srv *httptest.Server) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("new vault client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestDetect_NoDuplicates(t *testing.T) {
	stub := &stubReader{
		data: map[string]map[string]string{
			"secret/a": {"foo": "bar"},
			"secret/b": {"foo": "baz"},
		},
	}
	d := dedupe.New(stub, false)
	results, err := d.Detect(context.Background(), []string{"secret/a", "secret/b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no duplicates, got %d", len(results))
	}
}

func TestDetect_FindsDuplicates(t *testing.T) {
	stub := &stubReader{
		data: map[string]map[string]string{
			"secret/a": {"api_key": "supersecret"},
			"secret/b": {"api_key": "supersecret"},
			"secret/c": {"api_key": "different"},
		},
	}
	d := dedupe.New(stub, false)
	results, err := d.Detect(context.Background(), []string{"secret/a", "secret/b", "secret/c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 duplicate group, got %d", len(results))
	}
	if results[0].Key != "api_key" {
		t.Errorf("expected key api_key, got %q", results[0].Key)
	}
	if results[0].Value != "supersecret" {
		t.Errorf("expected value supersecret, got %q", results[0].Value)
	}
	if len(results[0].Paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(results[0].Paths))
	}
}

func TestDetect_MasksValues(t *testing.T) {
	stub := &stubReader{
		data: map[string]map[string]string{
			"secret/x": {"token": "abc123"},
			"secret/y": {"token": "abc123"},
		},
	}
	d := dedupe.New(stub, true)
	results, err := d.Detect(context.Background(), []string{"secret/x", "secret/y"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != "***" {
		t.Errorf("expected masked value, got %q", results[0].Value)
	}
}
