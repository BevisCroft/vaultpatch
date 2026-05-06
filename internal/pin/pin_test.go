package pin_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/pin"
	"github.com/your-org/vaultpatch/internal/vault"
)

func newMockVault(t *testing.T, handler http.Handler) *vault.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestPin_DryRun_NoWrites(t *testing.T) {
	written := false
	client := newMockVault(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			written = true
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	p := pin.New(client, "secret", true)
	r := p.Pin(context.Background(), "app/db", "alice", "freeze")
	if written {
		t.Error("expected no writes in dry-run mode")
	}
	if r.Err != nil {
		t.Errorf("unexpected error: %v", r.Err)
	}
	if !r.DryRun {
		t.Error("expected DryRun=true")
	}
}

func TestPin_LiveWritesMetadata(t *testing.T) {
	var capturedBody map[string]interface{}
	client := newMockVault(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	p := pin.New(client, "secret", false)
	r := p.Pin(context.Background(), "app/db", "bob", "hotfix")
	if r.Err != nil {
		t.Fatalf("Pin error: %v", r.Err)
	}
	cm, _ := capturedBody["custom_metadata"].(map[string]interface{})
	if cm["pinned"] != "true" {
		t.Errorf("expected pinned=true, got %v", cm["pinned"])
	}
	if cm["pinned_by"] != "bob" {
		t.Errorf("expected pinned_by=bob, got %v", cm["pinned_by"])
	}
}

func TestIsPinned_ReturnsTrueWhenPinned(t *testing.T) {
	client := newMockVault(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"custom_metadata": map[string]interface{}{
					"pinned": "true",
				},
			},
		})
	}))
	p := pin.New(client, "secret", false)
	ok, err := p.IsPinned(context.Background(), "app/db")
	if err != nil {
		t.Fatalf("IsPinned error: %v", err)
	}
	if !ok {
		t.Error("expected IsPinned=true")
	}
}

func TestFprint_PinResults(t *testing.T) {
	results := []pin.Result{
		{Path: "app/db", Op: "pin", DryRun: true},
		{Path: "app/cache", Op: "unpin", DryRun: false},
	}
	var buf bytes.Buffer
	pin.Fprint(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "app/db") {
		t.Error("expected app/db in output")
	}
	if !strings.Contains(out, "dry-run") {
		t.Error("expected dry-run tag in output")
	}
	if !strings.Contains(out, "app/cache") {
		t.Error("expected app/cache in output")
	}
}
