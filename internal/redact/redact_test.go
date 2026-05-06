package redact

import (
	"bytes"
	"strings"
	"testing"
)

func TestApply_NoMatch(t *testing.T) {
	r := New(DefaultRules())
	secrets := map[string]string{
		"database_host": "localhost",
		"port":          "5432",
	}
	out := r.Apply(secrets)
	if out["database_host"] != "localhost" {
		t.Errorf("expected unchanged value, got %q", out["database_host"])
	}
	if out["port"] != "5432" {
		t.Errorf("expected unchanged value, got %q", out["port"])
	}
}

func TestApply_MatchesDefaultRules(t *testing.T) {
	r := New(DefaultRules())
	secrets := map[string]string{
		"db_password": "supersecret",
		"api_token":   "tok-abc123",
		"host":        "example.com",
	}
	out := r.Apply(secrets)
	if out["db_password"] != "***REDACTED***" {
		t.Errorf("expected redacted password, got %q", out["db_password"])
	}
	if out["api_token"] != "***REDACTED***" {
		t.Errorf("expected redacted token, got %q", out["api_token"])
	}
	if out["host"] != "example.com" {
		t.Errorf("expected unchanged host, got %q", out["host"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	r := New(DefaultRules())
	orig := map[string]string{"password": "hunter2"}
	_ = r.Apply(orig)
	if orig["password"] != "hunter2" {
		t.Error("original map was mutated")
	}
}

func TestApply_CustomReplacement(t *testing.T) {
	rules := []Rule{{Pattern: "ssn", Replacement: "[SSN]"}}
	r := New(rules)
	out := r.Apply(map[string]string{"user_ssn": "123-45-6789"})
	if out["user_ssn"] != "[SSN]" {
		t.Errorf("expected [SSN], got %q", out["user_ssn"])
	}
}

func TestApply_CaseInsensitiveMatch(t *testing.T) {
	r := New(DefaultRules())
	out := r.Apply(map[string]string{"DB_PASSWORD": "val"})
	if out["DB_PASSWORD"] != "***REDACTED***" {
		t.Errorf("expected redacted, got %q", out["DB_PASSWORD"])
	}
}

func TestFprint_NoRedactions(t *testing.T) {
	var buf bytes.Buffer
	orig := map[string]string{"host": "localhost"}
	Fprint(&buf, orig, orig)
	if !strings.Contains(buf.String(), "no keys matched") {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestFprint_WithRedactions(t *testing.T) {
	var buf bytes.Buffer
	orig := map[string]string{"password": "secret", "host": "localhost"}
	redacted := map[string]string{"password": "***REDACTED***", "host": "localhost"}
	Fprint(&buf, orig, redacted)
	out := buf.String()
	if !strings.Contains(out, "1 key(s) redacted") {
		t.Errorf("expected count in output, got %q", out)
	}
	if !strings.Contains(out, "password") {
		t.Errorf("expected key name in output, got %q", out)
	}
}
