package protect_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/protect"
)

func results(rs ...protect.Result) []protect.Result { return rs }

func TestFprint_NoResults(t *testing.T) {
	var sb strings.Builder
	protect.Fprint(&sb, nil)
	if !strings.Contains(sb.String(), "no paths") {
		t.Errorf("expected 'no paths' message, got: %s", sb.String())
	}
}

func TestFprint_DryRunResult(t *testing.T) {
	var sb strings.Builder
	protect.Fprint(&sb, results(
		protect.Result{Path: "secret/prod/db", Protected: true, DryRun: true},
	))
	out := sb.String()
	if !strings.Contains(out, "dry-run") {
		t.Errorf("expected dry-run label, got: %s", out)
	}
	if !strings.Contains(out, "protected") {
		t.Errorf("expected 'protected' label, got: %s", out)
	}
}

func TestFprint_SuccessResult(t *testing.T) {
	var sb strings.Builder
	protect.Fprint(&sb, results(
		protect.Result{Path: "secret/prod/db", Protected: false},
	))
	out := sb.String()
	if !strings.Contains(out, "✓") {
		t.Errorf("expected success tick, got: %s", out)
	}
	if !strings.Contains(out, "unprotected") {
		t.Errorf("expected 'unprotected' label, got: %s", out)
	}
}

func TestFprint_ErrorResult(t *testing.T) {
	var sb strings.Builder
	protect.Fprint(&sb, results(
		protect.Result{Path: "secret/prod/db", Protected: true, Err: errors.New("permission denied")},
	))
	out := sb.String()
	if !strings.Contains(out, "✗") {
		t.Errorf("expected failure mark, got: %s", out)
	}
	if !strings.Contains(out, "permission denied") {
		t.Errorf("expected error message, got: %s", out)
	}
	if !strings.Contains(out, "0 succeeded, 1 failed") {
		t.Errorf("expected failure count, got: %s", out)
	}
}
