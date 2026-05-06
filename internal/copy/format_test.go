package copy_test

import (
	"strings"
	"testing"

	"errors"

	vaultcopy "github.com/yourusername/vaultpatch/internal/copy"
)

func TestFprint_NoResults(t *testing.T) {
	var sb strings.Builder
	vaultcopy.Fprint(&sb, nil)
	if !strings.Contains(sb.String(), "no copy operations") {
		t.Errorf("unexpected output: %q", sb.String())
	}
}

func TestFprint_DryRunResult(t *testing.T) {
	var sb strings.Builder
	vaultcopy.Fprint(&sb, []vaultcopy.Result{
		{Src: "secret/src", Dst: "secret/dst", DryRun: true},
	})
	out := sb.String()
	if !strings.Contains(out, "DRY-RUN") {
		t.Errorf("expected DRY-RUN in output, got: %q", out)
	}
}

func TestFprint_SuccessResult(t *testing.T) {
	var sb strings.Builder
	vaultcopy.Fprint(&sb, []vaultcopy.Result{
		{Src: "secret/src", Dst: "secret/dst"},
	})
	out := sb.String()
	if !strings.Contains(out, "COPIED") {
		t.Errorf("expected COPIED in output, got: %q", out)
	}
}

func TestFprint_ErrorResult(t *testing.T) {
	var sb strings.Builder
	vaultcopy.Fprint(&sb, []vaultcopy.Result{
		{Src: "secret/src", Dst: "secret/dst", Err: errors.New("permission denied")},
	})
	out := sb.String()
	if !strings.Contains(out, "ERROR") {
		t.Errorf("expected ERROR in output, got: %q", out)
	}
	if !strings.Contains(out, "permission denied") {
		t.Errorf("expected error message in output, got: %q", out)
	}
}
