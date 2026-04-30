package clone_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/your-org/vaultpatch/internal/audit"
	"github.com/your-org/vaultpatch/internal/clone"
)

type stubVault struct {
	data map[string]map[string]string
}

func (s *stubVault) ReadSecret(_ context.Context, path string) (map[string]string, error) {
	return s.data[path], nil
}

func (s *stubVault) WriteSecret(_ context.Context, path string, data map[string]string) error {
	if s.data == nil {
		s.data = make(map[string]map[string]string)
	}
	s.data[path] = data
	return nil
}

func newTestAuditor(t *testing.T) *audit.Auditor {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{}})
	}))
	t.Cleanup(srv.Close)
	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := vaultapi.NewClient(cfg)
	f, _ := os.CreateTemp(t.TempDir(), "audit-*.jsonl")
	t.Cleanup(func() { f.Close() })
	return audit.New(client, f, false)
}

func TestClone_DryRun_NoWrites(t *testing.T) {
	stub := &stubVault{
		data: map[string]map[string]string{
			"secret/src": {"key": "value"},
		},
	}
	a := newTestAuditor(t)
	c := clone.New(stub, stub, a, true)
	res := c.Clone(context.Background(), "secret/src", "secret/dst")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if _, ok := stub.data["secret/dst"]; ok {
		t.Error("dry-run should not write destination")
	}
	if res.Keys != 1 {
		t.Errorf("expected 1 key, got %d", res.Keys)
	}
}

func TestClone_LiveCopiesData(t *testing.T) {
	stub := &stubVault{
		data: map[string]map[string]string{
			"secret/src": {"alpha": "1", "beta": "2"},
		},
	}
	a := newTestAuditor(t)
	c := clone.New(stub, stub, a, false)
	res := c.Clone(context.Background(), "secret/src", "secret/dst")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if stub.data["secret/dst"]["alpha"] != "1" {
		t.Errorf("expected alpha=1 at destination")
	}
	if res.Keys != 2 {
		t.Errorf("expected 2 keys, got %d", res.Keys)
	}
}
