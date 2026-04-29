package watch_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/vaultpatch/internal/vault"
	"github.com/your-org/vaultpatch/internal/watch"
)

func newMockVault(t *testing.T, responses []map[string]interface{}) (*httptest.Server, *int64) {
	t.Helper()
	var callCount int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx := atomic.AddInt64(&callCount, 1) - 1
		if idx >= int64(len(responses)) {
			idx = int64(len(responses)) - 1
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": responses[idx]},
		})
	}))
	t.Cleanup(srv.Close)
	return srv, &callCount
}

func newTestClient(t *testing.T, addr string) *vault.Client {
	t.Helper()
	c, err := vault.NewClient(addr, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestWatch_EmitsEventOnChange(t *testing.T) {
	responses := []map[string]interface{}{
		{"key": "v1"},
		{"key": "v2"},
	}
	srv, _ := newMockVault(t, responses)
	client := newTestClient(t, srv.URL)

	w := watch.New(client, "secret", []string{"myapp/config"}, 10*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	ch := w.Watch(ctx)
	var events []watch.Event
	for e := range ch {
		events = append(events, e)
		if len(events) >= 1 {
			cancel()
		}
	}

	if len(events) == 0 {
		t.Fatal("expected at least one event, got none")
	}
	if events[0].Path != "myapp/config" {
		t.Errorf("unexpected path: %s", events[0].Path)
	}
	if len(events[0].Changes) == 0 {
		t.Error("expected changes in event")
	}
}

func TestWatch_NoEventWhenUnchanged(t *testing.T) {
	responses := []map[string]interface{}{
		{"key": "stable"},
		{"key": "stable"},
		{"key": "stable"},
	}
	srv, _ := newMockVault(t, responses)
	client := newTestClient(t, srv.URL)

	w := watch.New(client, "secret", []string{"myapp/config"}, 20*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	ch := w.Watch(ctx)
	var events []watch.Event
	for e := range ch {
		events = append(events, e)
	}

	if len(events) != 0 {
		t.Errorf("expected no events for unchanged secret, got %d", len(events))
	}
}
