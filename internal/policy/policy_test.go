package policy

import (
	"testing"
)

func basePolicy() *Policy {
	return &Policy{
		Name: "test-policy",
		Rules: []Rule{
			{PathPattern: "secret/prod/*", AllowedOps: []Op{OpRead}},
			{PathPattern: "secret/staging/*", AllowedOps: []Op{OpRead, OpWrite, OpDelete}},
			{PathPattern: "secret/shared/config", AllowedOps: []Op{OpRead, OpList}},
		},
	}
}

func TestCheck_NoViolations(t *testing.T) {
	p := basePolicy()
	reqs := map[string]Op{
		"secret/prod/db": OpRead,
		"secret/staging/api": OpWrite,
		"secret/shared/config": OpList,
	}
	got := p.Check(reqs)
	if len(got) != 0 {
		t.Fatalf("expected no violations, got %v", got)
	}
}

func TestCheck_WriteOnReadOnlyPath(t *testing.T) {
	p := basePolicy()
	reqs := map[string]Op{
		"secret/prod/db": OpWrite,
	}
	got := p.Check(reqs)
	if len(got) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(got))
	}
	if got[0].Path != "secret/prod/db" || got[0].Op != OpWrite {
		t.Errorf("unexpected violation: %v", got[0])
	}
}

func TestCheck_UnknownPath(t *testing.T) {
	p := basePolicy()
	reqs := map[string]Op{
		"secret/other/key": OpRead,
	}
	got := p.Check(reqs)
	if len(got) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(got))
	}
}

func TestCheck_MultipleViolations(t *testing.T) {
	p := basePolicy()
	reqs := map[string]Op{
		"secret/prod/db": OpDelete,
		"secret/prod/api": OpWrite,
	}
	got := p.Check(reqs)
	if len(got) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(got))
	}
}

func TestViolation_Error(t *testing.T) {
	v := Violation{Path: "secret/prod/db", Op: OpWrite}
	expected := `operation "write" not allowed on path "secret/prod/db"`
	if v.Error() != expected {
		t.Errorf("got %q, want %q", v.Error(), expected)
	}
}

func TestMatchPattern_ExactMatch(t *testing.T) {
	if !matchPattern("secret/shared/config", "secret/shared/config") {
		t.Error("expected exact match")
	}
	if matchPattern("secret/shared/config", "secret/shared/other") {
		t.Error("expected no match")
	}
}

func TestMatchPattern_Wildcard(t *testing.T) {
	if !matchPattern("secret/prod/*", "secret/prod/db") {
		t.Error("expected wildcard match")
	}
	if matchPattern("secret/prod/*", "secret/staging/db") {
		t.Error("expected no wildcard match")
	}
}
