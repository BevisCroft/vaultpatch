package scope_test

import (
	"bytes"
	"testing"

	"github.com/your-org/vaultpatch/internal/scope"
)

func TestNew_InvalidGlob(t *testing.T) {
	_, err := scope.New(scope.Config{Globs: []string{"[invalid"}})
	if err == nil {
		t.Fatal("expected error for invalid glob, got nil")
	}
}

func TestMatch_EmptyConfig_MatchesAll(t *testing.T) {
	f, err := scope.New(scope.Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range []string{"secret/a", "kv/b", "anything"} {
		if !f.Match(p) {
			t.Errorf("expected %q to match empty filter", p)
		}
	}
}

func TestMatch_Prefix(t *testing.T) {
	f, _ := scope.New(scope.Config{Prefixes: []string{"secret/prod/"}})
	if !f.Match("secret/prod/db") {
		t.Error("expected match for prefix")
	}
	if f.Match("secret/staging/db") {
		t.Error("expected no match outside prefix")
	}
}

func TestMatch_Glob(t *testing.T) {
	f, _ := scope.New(scope.Config{Globs: []string{"secret/*/password"}})
	if !f.Match("secret/prod/password") {
		t.Error("expected glob match")
	}
	if f.Match("secret/prod/token") {
		t.Error("expected no glob match")
	}
}

func TestMatch_ExactPath(t *testing.T) {
	f, _ := scope.New(scope.Config{Paths: []string{"secret/prod/api-key"}})
	if !f.Match("secret/prod/api-key") {
		t.Error("expected exact match")
	}
	if f.Match("secret/prod/api-key/extra") {
		t.Error("expected no match for longer path")
	}
}

func TestApply_ReturnsSortedSubset(t *testing.T) {
	f, _ := scope.New(scope.Config{Prefixes: []string{"kv/prod/"}})
	input := []string{"kv/prod/b", "kv/staging/x", "kv/prod/a"}
	got := f.Apply(input)
	want := []string{"kv/prod/a", "kv/prod/b"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestFprint_EmptyConfig(t *testing.T) {
	var buf bytes.Buffer
	scope.Fprint(&buf, scope.Config{})
	if buf.String() != "scope: (all paths)\n" {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestFprint_WithRules(t *testing.T) {
	var buf bytes.Buffer
	scope.Fprint(&buf, scope.Config{
		Prefixes: []string{"secret/prod/"},
		Globs:    []string{"kv/*/token"},
	})
	out := buf.String()
	for _, want := range []string{"scope:", "secret/prod/", "kv/*/token"} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}
