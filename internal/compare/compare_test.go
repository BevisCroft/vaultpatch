package compare

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/vault"
)

func newMockVault(t *testing.T, data map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/secret/data/")
		secrets, ok := data[path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": secrets}})
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

func TestCompare_NoDifferences(t *testing.T) {
	data := map[string]map[string]interface{}{"myapp/config": {"key": "val"}}
	srv := newMockVault(t, data)
	defer srv.Close()

	cl := newTestClient(t, srv)
	cmp := New("staging", cl, "prod", cl)
	res, err := cmp.Compare(context.Background(), "myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Differ) != 0 || len(res.OnlyIn["staging"]) != 0 || len(res.OnlyIn["prod"]) != 0 {
		t.Errorf("expected no differences, got %+v", res)
	}
}

func TestCompare_DetectsDifferences(t *testing.T) {
	srcSrv := newMockVault(t, map[string]map[string]interface{}{
		"app/cfg": {"shared": "old", "only_src": "x"},
	})
	defer srcSrv.Close()
	dstSrv := newMockVault(t, map[string]map[string]interface{}{
		"app/cfg": {"shared": "new", "only_dst": "y"},
	})
	defer dstSrv.Close()

	cmp := New("staging", newTestClient(t, srcSrv), "prod", newTestClient(t, dstSrv))
	res, err := cmp.Compare(context.Background(), "app/cfg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Differ) != 1 || res.Differ[0] != "shared" {
		t.Errorf("expected Differ=[shared], got %v", res.Differ)
	}
	if len(res.OnlyIn["staging"]) != 1 || res.OnlyIn["staging"][0] != "only_src" {
		t.Errorf("expected OnlyIn[staging]=[only_src], got %v", res.OnlyIn["staging"])
	}
	if len(res.OnlyIn["prod"]) != 1 || res.OnlyIn["prod"][0] != "only_dst" {
		t.Errorf("expected OnlyIn[prod]=[only_dst], got %v", res.OnlyIn["prod"])
	}
}

func TestFprint_NoDifferences(t *testing.T) {
	r := Result{Path: "app/cfg", OnlyIn: map[string][]string{"staging": {}, "prod": {}}, Differ: nil}
	var buf bytes.Buffer
	Fprint(&buf, r, "staging", "prod", true)
	if !strings.Contains(buf.String(), "no differences") {
		t.Errorf("expected 'no differences' in output, got: %s", buf.String())
	}
}

func TestFprint_WithDifferences(t *testing.T) {
	r := Result{
		Path:   "app/cfg",
		Differ: []string{"token"},
		OnlyIn: map[string][]string{"staging": {"extra"}, "prod": {}},
	}
	var buf bytes.Buffer
	Fprint(&buf, r, "staging", "prod", true)
	out := buf.String()
	if !strings.Contains(out, "token") {
		t.Errorf("expected 'token' in output")
	}
	if !strings.Contains(out, "extra") {
		t.Errorf("expected 'extra' in output")
	}
}
