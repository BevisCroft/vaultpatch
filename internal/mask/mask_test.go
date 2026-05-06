package mask_test

import (
	"testing"

	"github.com/your-org/vaultpatch/internal/mask"
)

func TestApply_NoMatch(t *testing.T) {
	m := mask.New(mask.DefaultRules())
	secrets := map[string]string{
		"username": "alice",
		"region":   "us-east-1",
	}
	got := m.Apply("app/config", secrets)
	if got["username"] != "alice" {
		t.Errorf("expected alice, got %s", got["username"])
	}
	if got["region"] != "us-east-1" {
		t.Errorf("expected us-east-1, got %s", got["region"])
	}
}

func TestApply_MatchesDefaultRules(t *testing.T) {
	m := mask.New(mask.DefaultRules())
	secrets := map[string]string{
		"password": "s3cr3t",
		"api_key":  "key-abc",
		"token":    "tok-xyz",
		"username": "bob",
	}
	got := m.Apply("service/prod", secrets)
	for _, k := range []string{"password", "api_key", "token"} {
		if got[k] != "***" {
			t.Errorf("key %q: expected ***, got %s", k, got[k])
		}
	}
	if got["username"] != "bob" {
		t.Errorf("username should not be masked")
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	m := mask.New(mask.DefaultRules())
	orig := map[string]string{"password": "real-value"}
	_ = m.Apply("any/path", orig)
	if orig["password"] != "real-value" {
		t.Error("original map was mutated")
	}
}

func TestApply_PathPatternFilters(t *testing.T) {
	import_re := mustCompile(t, `(?i)secret`)
	rules := []mask.Rule{
		{PathPattern: "internal/*", KeyPattern: import_re, Replacement: "[hidden]"},
	}
	m := mask.New(rules)

	masked := m.Apply("internal/db", map[string]string{"db_secret": "abc"})
	if masked["db_secret"] != "[hidden]" {
		t.Errorf("expected [hidden], got %s", masked["db_secret"])
	}

	notMasked := m.Apply("public/config", map[string]string{"db_secret": "abc"})
	if notMasked["db_secret"] != "abc" {
		t.Errorf("path outside pattern should not be masked")
	}
}

func TestApply_CustomReplacement(t *testing.T) {
	re := mustCompile(t, `private_key`)
	rules := []mask.Rule{
		{PathPattern: "*", KeyPattern: re, Replacement: "<REDACTED>"},
	}
	m := mask.New(rules)
	got := m.Apply("pki/certs", map[string]string{"private_key": "-----BEGIN"})
	if got["private_key"] != "<REDACTED>" {
		t.Errorf("unexpected replacement: %s", got["private_key"])
	}
}

func mustCompile(t *testing.T, pattern string) interface{ MatchString(string) bool } {
	t.Helper()
	import "regexp"
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("bad pattern %q: %v", pattern, err)
	}
	return re
}
