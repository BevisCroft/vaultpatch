package lock_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/your-org/vaultpatch/internal/lock"
)

// lockStore is a minimal in-memory KV v2 store used by the mock server.
type lockStore struct {
	mu   sync.Mutex
	data map[string]map[string]interface{}
}

func (s *lockStore) get(path string) map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data[path]
}

func (s *lockStore) set(path string, v map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[path] = v
}

func (s *lockStore) del(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, path)
}

func newMockVault(t *testing.T) (*httptest.Server, *vaultapi.Client) {
	t.Helper()
	store := &lockStore{data: make(map[string]map[string]interface{})}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/v1/"):]
		switch r.Method {
		case http.MethodPut, http.MethodPost:
			var body struct {
				Data map[string]interface{} `json:"data"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			store.set(path, body.Data)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": body.Data})
		case http.MethodGet:
			v := store.get(path)
			if v == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": v}})
		case http.MethodDelete:
			store.del(path)
			w.WriteHeader(http.StatusNoContent)
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("new vault client: %v", err)
	}
	client.SetToken("test-token")
	return srv, client
}

func TestAcquireAndRelease(t *testing.T) {
	_, client := newMockVault(t)
	l := lock.New(client, "secret", "owner-a")
	ctx := context.Background()

	if err := l.Acquire(ctx, "myapp/prod"); err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if err := l.Release(ctx, "myapp/prod"); err != nil {
		t.Fatalf("release: %v", err)
	}
}

func TestAcquire_BlockedByOtherOwner(t *testing.T) {
	_, client := newMockVault(t)
	ctx := context.Background()

	owner1 := lock.New(client, "secret", "owner-a")
	owner2 := lock.New(client, "secret", "owner-b")

	if err := owner1.Acquire(ctx, "myapp/prod"); err != nil {
		t.Fatalf("owner1 acquire: %v", err)
	}
	if err := owner2.Acquire(ctx, "myapp/prod"); err == nil {
		t.Fatal("expected error acquiring lock held by owner-a, got nil")
	}
}

func TestRelease_Noop_WhenNotOwner(t *testing.T) {
	_, client := newMockVault(t)
	ctx := context.Background()

	owner1 := lock.New(client, "secret", "owner-a")
	owner2 := lock.New(client, "secret", "owner-b")

	if err := owner1.Acquire(ctx, "myapp/staging"); err != nil {
		t.Fatalf("acquire: %v", err)
	}
	// owner2 releasing a lock it doesn't own should be a no-op, not an error.
	if err := owner2.Release(ctx, "myapp/staging"); err != nil {
		t.Fatalf("unexpected error on no-op release: %v", err)
	}
}
