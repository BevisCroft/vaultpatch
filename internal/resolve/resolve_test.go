package resolve_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/seatgeek/vaultpatch/internal/resolve"
)

type stubReader struct {
	data map[string]map[string]string
	err  error
}

func (s *stubReader) ReadSecret(_ context.Context, path string) (map[string]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	if m, ok := s.data[path]; ok {
		return m, nil
	}
	return nil, errors.New("not found")
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

func TestApply_NoReferences(t *testing.T) {
	r := resolve.New(&stubReader{data: nil})
	secrets := map[string]string{"foo": "bar", "baz": "qux"}
	out, results := r.Apply(context.Background(), secrets)
	if out["foo"] != "bar" || out["baz"] != "qux" {
		t.Errorf("unexpected output: %v", out)
	}
	if len(results) != 0 {
		t.Errorf("expected no results, got %d", len(results))
	}
}

func TestApply_ResolvesReference(t *testing.T) {
	stub := &stubReader{
		data: map[string]map[string]string{
			"secret/db": {"password": "s3cr3t"},
		},
	}
	r := resolve.New(stub)
	secrets := map[string]string{"db_pass": "{{secret/db:password}}"}
	out, results := r.Apply(context.Background(), secrets)
	if out["db_pass"] != "s3cr3t" {
		t.Errorf("expected 's3cr3t', got %q", out["db_pass"])
	}
	if len(results) != 1 || results[0].Err != nil {
		t.Errorf("unexpected results: %+v", results)
	}
}

func TestApply_MissingKey(t *testing.T) {
	stub := &stubReader{
		data: map[string]map[string]string{
			"secret/db": {"user": "admin"},
		},
	}
	r := resolve.New(stub)
	secrets := map[string]string{"val": "{{secret/db:password}}"}
	out, results := r.Apply(context.Background(), secrets)
	if out["val"] != "{{secret/db:password}}" {
		t.Errorf("placeholder should be preserved, got %q", out["val"])
	}
	if len(results) != 1 || results[0].Err == nil {
		t.Errorf("expected error result, got %+v", results)
	}
}

func TestApply_ReadError(t *testing.T) {
	stub := &stubReader{err: errors.New("connection refused")}
	r := resolve.New(stub)
	secrets := map[string]string{"x": "{{secret/db:password}}"}
	out, results := r.Apply(context.Background(), secrets)
	if !strings.Contains(out["x"], "{{") {
		t.Errorf("expected placeholder preserved, got %q", out["x"])
	}
	if len(results) != 1 || results[0].Err == nil {
		t.Errorf("expected error result")
	}
}

func TestApply_MultipleRefsInValue(t *testing.T) {
	stub := &stubReader{
		data: map[string]map[string]string{
			"secret/a": {"host": "localhost"},
			"secret/b": {"port": "5432"},
		},
	}
	r := resolve.New(stub)
	secrets := map[string]string{"dsn": "{{secret/a:host}}:{{secret/b:port}}"}
	out, results := r.Apply(context.Background(), secrets)
	if out["dsn"] != "localhost:5432" {
		t.Errorf("expected 'localhost:5432', got %q", out["dsn"])
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

var _ = newTestClient // suppress unused warning; kept for future HTTP-level tests
var _ = http.HandlerFunc(nil)
