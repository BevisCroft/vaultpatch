package rollback_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/user/vaultpatch/internal/rollback"
	"github.com/user/vaultpatch/internal/snapshot"
)

// fakeWriter records calls and optionally returns an error.
type fakeWriter struct {
	writes map[string]map[string]interface{}
	failOn string
}

func (f *fakeWriter) WriteSecret(_ context.Context, path string, data map[string]interface{}) error {
	if f.failOn == path {
		return errors.New("write failed")
	}
	f.writes[path] = data
	return nil
}

func newTestSnap() *snapshot.Snapshot {
	return &snapshot.Snapshot{
		CapturedAt: time.Now(),
		Secrets: map[string]map[string]interface{}{
			"secret/app/db": {"password": "old-pass"},
			"secret/app/api": {"key": "old-key"},
		},
	}
}

func newMockAuditServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusNoContent)
	}))
}

func TestApply_DryRun_NoWrites(t *testing.T) {
	fw := &fakeWriter{writes: make(map[string]map[string]interface{})}
	applier := rollback.New(fw, nil, true)

	results, err := applier.Apply(context.Background(), newTestSnap())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fw.writes) != 0 {
		t.Errorf("expected no writes in dry-run, got %d", len(fw.writes))
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestApply_WritesAllPaths(t *testing.T) {
	fw := &fakeWriter{writes: make(map[string]map[string]interface{})}
	applier := rollback.New(fw, nil, false)

	_, err := applier.Apply(context.Background(), newTestSnap())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fw.writes) != 2 {
		t.Errorf("expected 2 writes, got %d", len(fw.writes))
	}
}

func TestApply_NilSnapshot(t *testing.T) {
	fw := &fakeWriter{writes: make(map[string]map[string]interface{})}
	applier := rollback.New(fw, nil, false)

	_, err := applier.Apply(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "snapshot must not be nil") {
		t.Errorf("expected nil snapshot error, got %v", err)
	}
}

func TestApply_RecordsWriteError(t *testing.T) {
	fw := &fakeWriter{writes: make(map[string]map[string]interface{}), failOn: "secret/app/db"}
	applier := rollback.New(fw, nil, false)

	results, err := applier.Apply(context.Background(), newTestSnap())
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	var failed int
	for _, r := range results {
		if r.Err != nil {
			failed++
		}
	}
	if failed != 1 {
		t.Errorf("expected 1 failed result, got %d", failed)
	}
}
