package quota_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/yourusername/vaultpatch/internal/quota"
)

func newMockVault(t *testing.T, paths map[string][]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/")
		path = strings.TrimSuffix(path, "/")
		keys, ok := paths[path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body := map[string]any{"data": map[string]any{"keys": keys}}
		_ = json.NewEncoder(w).Encode(body)
	}))
}

func newTestClient(t *testing.T, srv *httptest.Server) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("vault client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

type stubLister struct{ data map[string][]string }

func (s *stubLister) ListSecrets(_ context.Context, path string) ([]string, error) {
	keys, ok := s.data[path]
	if !ok {
		return nil, nil
	}
	return keys, nil
}

func TestCheck_WithinLimit(t *testing.T) {
	lister := &stubLister{data: map[string][]string{
		"secret/dev": {"a", "b", "c"},
	}}
	checker := quota.New(lister, []quota.Rule{{Path: "secret/dev", Limit: 5}})
	results := checker.Check(context.Background())
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Exceeds {
		t.Error("expected no quota violation")
	}
	if results[0].Current != 3 {
		t.Errorf("expected current=3, got %d", results[0].Current)
	}
}

func TestCheck_ExceedsLimit(t *testing.T) {
	lister := &stubLister{data: map[string][]string{
		"secret/prod": {"x", "y", "z", "w"},
	}}
	checker := quota.New(lister, []quota.Rule{{Path: "secret/prod", Limit: 2}})
	results := checker.Check(context.Background())
	if !results[0].Exceeds {
		t.Error("expected quota violation")
	}
	if !quota.AnyExceeds(results) {
		t.Error("AnyExceeds should return true")
	}
}

func TestFprint_Output(t *testing.T) {
	results := []quota.Result{
		{Path: "secret/dev", Limit: 10, Current: 3, Exceeds: false},
		{Path: "secret/prod", Limit: 2, Current: 5, Exceeds: true},
	}
	var buf bytes.Buffer
	quota.Fprint(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "OK") {
		t.Error("expected OK status in output")
	}
	if !strings.Contains(out, "EXCEEDS") {
		t.Error("expected EXCEEDS status in output")
	}
}

func TestFprint_Empty(t *testing.T) {
	var buf bytes.Buffer
	quota.Fprint(&buf, nil)
	if !strings.Contains(buf.String(), "no quota rules") {
		t.Error("expected empty message")
	}
}
