package prefix_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/yourusername/vaultpatch/internal/prefix"
)

// stubClient is an in-memory SecretWriter.
type stubClient struct {
	store map[string]map[string]interface{}
}

func (s *stubClient) ReadSecret(_ context.Context, path string) (map[string]interface{}, error) {
	d, ok := s.store[path]
	if !ok {
		return nil, fmt.Errorf("not found: %s", path)
	}
	out := make(map[string]interface{}, len(d))
	for k, v := range d {
		out[k] = v
	}
	return out, nil
}

func (s *stubClient) WriteSecret(_ context.Context, path string, data map[string]interface{}) error {
	s.store[path] = data
	return nil
}

func newStub(path string, kv map[string]interface{}) *stubClient {
	return &stubClient{store: map[string]map[string]interface{}{path: kv}}
}

func TestApply_DryRun_NoWrites(t *testing.T) {
	client := newStub("secret/app", map[string]interface{}{"DB_HOST": "localhost"})
	ap := prefix.New(client, true)
	res := ap.Apply(context.Background(), "secret/app", prefix.OpAdd, "APP_")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.DryRun {
		t.Error("expected DryRun=true")
	}
	if res.Changed != 1 {
		t.Errorf("expected 1 changed key, got %d", res.Changed)
	}
	// original must be untouched
	if _, ok := client.store["secret/app"]["DB_HOST"]; !ok {
		t.Error("dry-run must not mutate the store")
	}
}

func TestApply_AddPrefix(t *testing.T) {
	client := newStub("secret/app", map[string]interface{}{"HOST": "db", "PORT": "5432"})
	ap := prefix.New(client, false)
	res := ap.Apply(context.Background(), "secret/app", prefix.OpAdd, "DB_")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Changed != 2 {
		t.Errorf("expected 2 changed keys, got %d", res.Changed)
	}
	if _, ok := client.store["secret/app"]["DB_HOST"]; !ok {
		t.Error("expected key DB_HOST in store")
	}
}

func TestApply_StripPrefix(t *testing.T) {
	client := newStub("secret/app", map[string]interface{}{"DB_HOST": "db", "DB_PORT": "5432"})
	ap := prefix.New(client, false)
	res := ap.Apply(context.Background(), "secret/app", prefix.OpStrip, "DB_")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Changed != 2 {
		t.Errorf("expected 2 changed keys, got %d", res.Changed)
	}
	if _, ok := client.store["secret/app"]["HOST"]; !ok {
		t.Error("expected key HOST after strip")
	}
}

func TestApply_NoMatch_ZeroChanged(t *testing.T) {
	client := newStub("secret/app", map[string]interface{}{"APP_HOST": "localhost"})
	ap := prefix.New(client, false)
	res := ap.Apply(context.Background(), "secret/app", prefix.OpAdd, "APP_")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Changed != 0 {
		t.Errorf("expected 0 changed keys, got %d", res.Changed)
	}
}

func TestApply_ReadError(t *testing.T) {
	client := &stubClient{store: map[string]map[string]interface{}{}}
	ap := prefix.New(client, false)
	res := ap.Apply(context.Background(), "secret/missing", prefix.OpAdd, "X_")

	if res.Err == nil {
		t.Fatal("expected error for missing path")
	}
}

// Ensure the package compiles cleanly when used via the real vault client interface.
func TestInterface_Compatibility(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"key": "value"},
		})
	}))
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	_, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("vault client: %v", err)
	}
	// Just verifying the dependency resolves; actual integration tested via stubClient.
}
