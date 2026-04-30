package rename_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/vaultpatch/internal/rename"
)

// stubClient is a simple in-memory VaultClient for testing.
type stubClient struct {
	store   map[string]map[string]interface{}
	deleted []string
	readErr error
	writeErr error
}

func newStub(data map[string]map[string]interface{}) *stubClient {
	if data == nil {
		data = map[string]map[string]interface{}{}
	}
	return &stubClient{store: data}
}

func (s *stubClient) ReadSecret(_ context.Context, path string) (map[string]interface{}, error) {
	if s.readErr != nil {
		return nil, s.readErr
	}
	v, ok := s.store[path]
	if !ok {
		return nil, errors.New("not found")
	}
	return v, nil
}

func (s *stubClient) WriteSecret(_ context.Context, path string, data map[string]interface{}) error {
	if s.writeErr != nil {
		return s.writeErr
	}
	s.store[path] = data
	return nil
}

func (s *stubClient) DeleteSecret(_ context.Context, path string) error {
	delete(s.store, path)
	s.deleted = append(s.deleted, path)
	return nil
}

// newMockVaultServer satisfies the build; unused here but keeps parity with other test files.
func newMockVaultServer(_ *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

func TestRename_Success(t *testing.T) {
	client := newStub(map[string]map[string]interface{}{
		"secret/old": {"key": "value"},
	})
	r := rename.New(client, nil, false)
	res := r.Rename(context.Background(), "secret/old", "secret/new")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if _, ok := client.store["secret/new"]; !ok {
		t.Error("expected secret/new to exist")
	}
	if _, ok := client.store["secret/old"]; ok {
		t.Error("expected secret/old to be deleted")
	}
}

func TestRename_DryRun_NoMutations(t *testing.T) {
	client := newStub(map[string]map[string]interface{}{
		"secret/old": {"key": "value"},
	})
	r := rename.New(client, nil, true)
	res := r.Rename(context.Background(), "secret/old", "secret/new")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.DryRun {
		t.Error("expected DryRun flag to be true")
	}
	if _, ok := client.store["secret/new"]; ok {
		t.Error("dry-run should not write secret/new")
	}
	if _, ok := client.store["secret/old"]; !ok {
		t.Error("dry-run should not delete secret/old")
	}
}

func TestRename_ReadError(t *testing.T) {
	client := newStub(nil)
	client.readErr = errors.New("permission denied")
	r := rename.New(client, nil, false)
	res := r.Rename(context.Background(), "secret/missing", "secret/new")

	if res.Err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFprint_DryRun(t *testing.T) {
	var buf bytes.Buffer
	res := rename.Result{Src: "a", Dst: "b", DryRun: true}
	rename.Fprint(&buf, res)
	if got := buf.String(); got == "" {
		t.Error("expected non-empty output")
	}
}
