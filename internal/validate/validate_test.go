package validate_test

import (
	"testing"

	"github.com/your-org/vaultpatch/internal/diff"
	"github.com/your-org/vaultpatch/internal/validate"
)

func TestValidate_NoIssues(t *testing.T) {
	changes := []diff.Change{
		{Op: diff.OpAdd, Path: "secret/app", Key: "DB_HOST", NewValue: "localhost"},
	}
	res := validate.Validate(changes)
	if !res.OK() {
		t.Fatalf("expected OK, got issues: %v", res.Issues)
	}
	if len(res.Issues) != 0 {
		t.Fatalf("expected 0 issues, got %d", len(res.Issues))
	}
}

func TestValidate_EmptyValueWarning(t *testing.T) {
	changes := []diff.Change{
		{Op: diff.OpAdd, Path: "secret/app", Key: "API_KEY", NewValue: ""},
	}
	res := validate.Validate(changes)
	if !res.OK() {
		t.Fatal("expected OK (warnings do not fail validation)")
	}
	if len(res.Issues) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(res.Issues))
	}
	if res.Issues[0].Severity != "warning" {
		t.Errorf("expected warning, got %s", res.Issues[0].Severity)
	}
}

func TestValidate_WhitespaceKeyError(t *testing.T) {
	changes := []diff.Change{
		{Op: diff.OpAdd, Path: "secret/app", Key: "BAD KEY", NewValue: "v"},
	}
	res := validate.Validate(changes)
	if res.OK() {
		t.Fatal("expected validation failure for whitespace key")
	}
	if len(res.Issues) != 1 || res.Issues[0].Severity != "error" {
		t.Errorf("unexpected issues: %v", res.Issues)
	}
}

func TestValidate_RemoveRequiredKeyError(t *testing.T) {
	changes := []diff.Change{
		{Op: diff.OpRemove, Path: "secret/app", Key: "DB_PASS_required", OldValue: "secret"},
	}
	res := validate.Validate(changes)
	if res.OK() {
		t.Fatal("expected validation failure for removing _required key")
	}
}

func TestValidate_MultipleIssues(t *testing.T) {
	changes := []diff.Change{
		{Op: diff.OpAdd, Path: "secret/app", Key: "GOOD", NewValue: ""},
		{Op: diff.OpRemove, Path: "secret/app", Key: "TOKEN_required", OldValue: "x"},
	}
	res := validate.Validate(changes)
	if res.OK() {
		t.Fatal("expected overall failure")
	}
	if len(res.Issues) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(res.Issues))
	}
}
