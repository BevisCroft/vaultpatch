package expire

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/youorg/vaultpatch/internal/vault"
)

func newMockVault(t *testing.T, data map[string]map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/")
		secrets, ok := data[path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": secrets})
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

func TestCheck_NoExpiryMetadata(t *testing.T) {
	srv := newMockVault(t, map[string]map[string]string{
		"secret/myapp": {"API_KEY": "abc123"},
	})
	defer srv.Close()
	checker := New(newTestClient(t, srv), "secret")
	results, err := checker.Check(context.Background(), []string{"myapp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestCheck_ExpiredSecret(t *testing.T) {
	past := time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339)
	srv := newMockVault(t, map[string]map[string]string{
		"secret/myapp": {"_expires_at": past, "_expire_note": "rotate asap"},
	})
	defer srv.Close()
	checker := New(newTestClient(t, srv), "secret")
	results, err := checker.Check(context.Background(), []string{"myapp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Expired {
		t.Error("expected secret to be expired")
	}
	if results[0].Note != "rotate asap" {
		t.Errorf("unexpected note: %q", results[0].Note)
	}
}

func TestCheck_ValidSecret(t *testing.T) {
	future := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	srv := newMockVault(t, map[string]map[string]string{
		"secret/creds": {"_expires_at": future},
	})
	defer srv.Close()
	checker := New(newTestClient(t, srv), "secret")
	results, err := checker.Check(context.Background(), []string{"creds"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Expired {
		t.Error("expected secret to be valid")
	}
	if results[0].DaysLeft < 28 {
		t.Errorf("expected ~30 days left, got %d", results[0].DaysLeft)
	}
}

func TestFprint_ExpiredOutput(t *testing.T) {
	results := []Result{
		{Entry: Entry{Path: "secret/old", ExpiresAt: time.Now().Add(-1 * time.Hour)}, Expired: true, DaysLeft: -1},
	}
	var buf bytes.Buffer
	Fprint(&buf, results)
	if !strings.Contains(buf.String(), "EXPIRED") {
		t.Errorf("expected EXPIRED in output, got: %s", buf.String())
	}
}

func TestFprint_NoResults(t *testing.T) {
	var buf bytes.Buffer
	Fprint(&buf, nil)
	if !strings.Contains(buf.String(), "no expiry metadata") {
		t.Errorf("unexpected output: %s", buf.String())
	}
}
