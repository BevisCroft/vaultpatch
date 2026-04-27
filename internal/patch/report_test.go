package patch_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/example/vaultpatch/internal/diff"
	"github.com/example/vaultpatch/internal/patch"
)

func TestFprintReport_DryRun(t *testing.T) {
	results := []patch.Result{
		{Path: "secret/app", Key: "TOKEN", Op: diff.OpAdd, Success: true},
		{Path: "secret/app", Key: "OLD", Op: diff.OpRemove, Success: true},
	}

	var buf bytes.Buffer
	patch.FprintReport(&buf, results, true)
	out := buf.String()

	if !strings.Contains(out, "Dry-run") {
		t.Error("expected 'Dry-run' in output")
	}
	if !strings.Contains(out, "2 change(s)") {
		t.Errorf("expected change count in output, got:\n%s", out)
	}
	if !strings.Contains(out, "~") {
		t.Error("expected dry-run symbol '~' in output")
	}
}

func TestFprintReport_WithFailure(t *testing.T) {
	results := []patch.Result{
		{Path: "secret/db", Key: "PASS", Op: diff.OpUpdate, Success: true},
		{Path: "secret/db", Key: "USER", Op: diff.OpAdd, Success: false, Err: fmt.Errorf("permission denied")},
	}

	var buf bytes.Buffer
	patch.FprintReport(&buf, results, false)
	out := buf.String()

	if !strings.Contains(out, "1 ok, 1 failed") {
		t.Errorf("expected failure summary, got:\n%s", out)
	}
	if !strings.Contains(out, "permission denied") {
		t.Error("expected error message in output")
	}
}
