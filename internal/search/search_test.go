package search_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/search"
	"github.com/your-org/vaultpatch/internal/vault"
)

func newMockVault(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.RawQuery, "list=true"):
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"keys": []string{"app/config", "app/db"}},
			})
		case strings.HasSuffix(r.URL.Path, "/app/config"):
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"api_key": "secret123", "region": "us-east-1"},
			})
		case strings.HasSuffix(r.URL.Path, "/app/db"):
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"db_password": "hunter2", "db_host": "localhost"},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
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

func TestFind_MatchesKey(t *testing.T) {
	srv := newMockVault(t)
	defer srv.Close()

	s := search.New(newTestClient(t, srv), "secret")
	results, err := s.Find(context.Background(), "api_key", false)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Key != "api_key" {
		t.Errorf("expected key api_key, got %s", results[0].Key)
	}
}

func TestFind_MatchesValue(t *testing.T) {
	srv := newMockVault(t)
	defer srv.Close()

	s := search.New(newTestClient(t, srv), "secret")
	results, err := s.Find(context.Background(), "hunter2", true)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if len(results) != 1 || results[0].Key != "db_password" {
		t.Errorf("expected db_password match, got %+v", results)
	}
}

func TestFind_NoMatch(t *testing.T) {
	srv := newMockVault(t)
	defer srv.Close()

	s := search.New(newTestClient(t, srv), "secret")
	results, err := s.Find(context.Background(), "nonexistent", true)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestFprint_MasksValues(t *testing.T) {
	results := []search.Result{
		{Path: "secret/app/config", Key: "api_key", Value: "topsecret"},
	}
	var buf bytes.Buffer
	search.Fprint(&buf, results, true)
	if strings.Contains(buf.String(), "topsecret") {
		t.Error("expected value to be masked")
	}
	if !strings.Contains(buf.String(), "***") {
		t.Error("expected mask placeholder in output")
	}
}

func TestFprint_NoResults(t *testing.T) {
	var buf bytes.Buffer
	search.Fprint(&buf, nil, false)
	if !strings.Contains(buf.String(), "no matches found") {
		t.Errorf("expected no-match message, got: %s", buf.String())
	}
}
