package revert_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/youorg/vaultpatch/internal/audit"
	"github.com/youorg/vaultpatch/internal/revert"
	"github.com/youorg/vaultpatch/internal/snapshot"
	"github.com/youorg/vaultpatch/internal/vault"
)

func newMockVault(t *testing.T, secrets map[string]map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func newTestClient(t *testing.T, srv *httptest.Server) *vault.Client {
	t.Helper()
	c, err := vault.NewClient(vault.Config{Address: srv.URL, Token: "test-token", Mount: "secret"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func newTestAuditor(t *testing.T) *audit.Auditor {
	t.Helper()
	var buf bytes.Buffer
	return audit.New(&buf)
}

func makeSnap(secrets map[string]map[string]string) *snapshot.Snapshot {
	s := &snapshot.Snapshot{
		Secrets:   make(map[string]map[string]string),
		CapturedAt: time.Now().UTC(),
	}
	for k, v := range secrets {
		s.Secrets[k] = v
	}
	return s
}

func TestApply_DryRun_NoWrites(t *testing.T) {
	writes := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			writes++
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	client := newTestClient(t, srv)
	auditor := newTestAuditor(t)
	snap := makeSnap(map[string]map[string]string{
		"secret/data/app": {"key": "val"},
	})

	r := revert.New(client, auditor, true)
	results := r.Apply(context.Background(), snap, []string{"secret/data/app"})

	if writes != 0 {
		t.Errorf("expected 0 writes in dry-run, got %d", writes)
	}
	if len(results) != 1 || results[0].DryRun != true {
		t.Errorf("expected 1 dry-run result")
	}
}

func TestApply_SkipsPathNotInSnapshot(t *testing.T) {
	srv := newMockVault(t, nil)
	t.Cleanup(srv.Close)

	client := newTestClient(t, srv)
	auditor := newTestAuditor(t)
	snap := makeSnap(map[string]map[string]string{})

	r := revert.New(client, auditor, false)
	results := r.Apply(context.Background(), snap, []string{"secret/data/missing"})

	if len(results) != 1 || !results[0].Skipped {
		t.Errorf("expected skipped result for path not in snapshot")
	}
}

func TestFprint_Summary(t *testing.T) {
	results := []revert.Result{
		{Path: "secret/data/a", DryRun: false},
		{Path: "secret/data/b", Skipped: true},
		{Path: "secret/data/c", Err: fmt.Errorf("write failed")},
	}
	var buf bytes.Buffer
	revert.Fprint(&buf, results)
	out := buf.String()

	if !strings.Contains(out, "✓") {
		t.Error("expected success marker")
	}
	if !strings.Contains(out, "skipped") {
		t.Error("expected skipped in output")
	}
	if !strings.Contains(out, "errors") {
		t.Error("expected errors in output")
	}
}

func init() {
	// ensure json import used in potential future marshal helpers
	_ = json.Marshal
	_ = fmt.Sprintf
}
