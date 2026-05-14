package scope_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/scope"
)

func TestFprint_PrefixOnly(t *testing.T) {
	var buf bytes.Buffer
	scope.Fprint(&buf, scope.Config{Prefixes: []string{"secret/prod/", "secret/shared/"}})
	out := buf.String()
	if !strings.Contains(out, "prefixes") {
		t.Errorf("expected 'prefixes' in output, got:\n%s", out)
	}
	if strings.Contains(out, "globs") {
		t.Errorf("unexpected 'globs' in output:\n%s", out)
	}
	if strings.Contains(out, "paths") {
		t.Errorf("unexpected 'paths' in output:\n%s", out)
	}
}

func TestFprint_AllRules(t *testing.T) {
	var buf bytes.Buffer
	scope.Fprint(&buf, scope.Config{
		Prefixes: []string{"secret/prod/"},
		Globs:    []string{"kv/*/db"},
		Paths:    []string{"secret/shared/key"},
	})
	out := buf.String()
	for _, want := range []string{"prefixes", "globs", "paths", "secret/prod/", "kv/*/db", "secret/shared/key"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output:\n%s", want, out)
		}
	}
}
