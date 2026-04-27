package promote_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/promote"
	"github.com/your-org/vaultpatch/internal/vault"
)

func newMockVault(t *testing.T, secrets map[string]map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/secret/data/")
		listPath := strings.TrimPrefix(r.URL.Path, "/v1/secret/metadata/")

		switch {
		case r.Method == http.MethodGet && r.URL.Query().Get("list") == "true":
			var keys []string
			for k := range secrets {
				if strings.HasPrefix(k, listPath) {
					keys = append(keys, k)
				}
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"keys": keys}})
		case r.Method == http.MethodGet:
			data, ok := secrets[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": data}})
		case r.Method == http.MethodPost || r.Method == http.MethodPut:
			var body struct {
				Data map[string]string `json:"data"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			secrets[path] = body.Data
			w.WriteHeader(http.StatusOK)
		}
	}))
}

func newTestClient(t *testing.T, srv *httptest.Server) *vault.Client {
	t.Helper()
	c, err := vault.NewClient(srv.URL, "test-token", "secret")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestPromote_DryRun_NoWrites(t *testing.T) {
	srcSecrets := map[string]map[string]string{"app/db": {"password": "s3cr3t"}}
	dstSecrets := map[string]map[string]string{}

	srcSrv := newMockVault(t, srcSecrets)
	dstSrv := newMockVault(t, dstSecrets)
	defer srcSrv.Close()
	defer dstSrv.Close()

	p := promote.New(newTestClient(t, srcSrv), newTestClient(t, dstSrv), promote.Options{DryRun: true})
	results, err := p.Promote(context.Background(), "app")
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if len(dstSecrets) != 0 {
		t.Error("dry-run should not write to destination")
	}
}

func TestPromote_Overwrite(t *testing.T) {
	srcSecrets := map[string]map[string]string{"app/db": {"password": "new"}}
	dstSecrets := map[string]map[string]string{"app/db": {"password": "old", "host": "localhost"}}

	srcSrv := newMockVault(t, srcSecrets)
	dstSrv := newMockVault(t, dstSecrets)
	defer srcSrv.Close()
	defer dstSrv.Close()

	p := promote.New(newTestClient(t, srcSrv), newTestClient(t, dstSrv), promote.Options{Overwrite: true})
	_, err := p.Promote(context.Background(), "app")
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if got := dstSecrets["app/db"]["password"]; got != "new" {
		t.Errorf("password = %q, want %q", got, "new")
	}
	if got := dstSecrets["app/db"]["host"]; got != "localhost" {
		t.Errorf("host = %q, want %q", got, "localhost")
	}
}
