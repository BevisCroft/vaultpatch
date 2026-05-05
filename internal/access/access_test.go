package access_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/example/vaultpatch/internal/access"
)

func baseRules() []access.Rule {
	return []access.Rule{
		{PathPrefix: "secret/prod/", Permissions: []access.Permission{access.PermRead}},
		{PathPrefix: "secret/dev/", Permissions: []access.Permission{access.PermRead, access.PermWrite, access.PermDelete, access.PermList}},
	}
}

func TestCheck_AllowedRead(t *testing.T) {
	c := access.New(baseRules())
	r := c.Check(context.Background(), "secret/prod/db", access.PermRead)
	if !r.Allowed {
		t.Fatalf("expected allowed, got denied: %s", r.Reason)
	}
}

func TestCheck_DeniedWrite_ReadOnlyPrefix(t *testing.T) {
	c := access.New(baseRules())
	r := c.Check(context.Background(), "secret/prod/db", access.PermWrite)
	if r.Allowed {
		t.Fatal("expected denied, got allowed")
	}
	if !strings.Contains(r.Reason, "not in rule") {
		t.Errorf("unexpected reason: %s", r.Reason)
	}
}

func TestCheck_NoMatchingRule(t *testing.T) {
	c := access.New(baseRules())
	r := c.Check(context.Background(), "secret/staging/app", access.PermRead)
	if r.Allowed {
		t.Fatal("expected denied for unmatched path")
	}
	if !strings.Contains(r.Reason, "no rule matched") {
		t.Errorf("unexpected reason: %s", r.Reason)
	}
}

func TestCheckAll_MixedResults(t *testing.T) {
	c := access.New(baseRules())
	checks := []struct {
		Path string
		Op   access.Permission
	}{
		{"secret/dev/svc", access.PermWrite},
		{"secret/prod/svc", access.PermDelete},
	}
	results := c.CheckAll(context.Background(), checks)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].Allowed {
		t.Error("expected dev/svc write to be allowed")
	}
	if results[1].Allowed {
		t.Error("expected prod/svc delete to be denied")
	}
}

// TestCheckAll_EmptyChecks verifies that CheckAll returns an empty slice
// rather than nil when called with no checks.
func TestCheckAll_EmptyChecks(t *testing.T) {
	c := access.New(baseRules())
	results := c.CheckAll(context.Background(), nil)
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestFprint_Output(t *testing.T) {
	results := []access.Result{
		{Path: "secret/prod/db", Op: access.PermRead, Allowed: true, Reason: "matched rule prefix \"secret/prod/\""},
		{Path: "secret/prod/db", Op: access.PermWrite, Allowed: false, Reason: "operation not in rule"},
	}
	var buf bytes.Buffer
	access.Fprint(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "1/2 allowed") {
		t.Errorf("expected summary line, got: %s", out)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("expected allow symbol in output")
	}
	if !strings.Contains(out, "✗") {
		t.Errorf("expected deny symbol in output")
	}
}

func TestFprint_Empty(t *testing.T) {
	var buf bytes.Buffer
	access.Fprint(&buf, nil)
	if !strings.Contains(buf.String(), "no access checks") {
		t.Errorf("expected empty message")
	}
}
