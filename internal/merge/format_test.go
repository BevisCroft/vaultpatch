package merge_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/merge"
)

func TestFprint_NoResults(t *testing.T) {
	var buf bytes.Buffer
	merge.Fprint(&buf, nil)
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestFprint_SuccessResult(t *testing.T) {
	var buf bytes.Buffer
	results := []merge.Result{
		{Path: "secret/dst", Merged: map[string]string{"a": "1", "b": "2"}},
	}
	merge.Fprint(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "secret/dst") {
		t.Errorf("expected path in output, got %q", out)
	}
	if !strings.Contains(out, "keys=2") {
		t.Errorf("expected keys=2 in output, got %q", out)
	}
	if !strings.Contains(out, "merged") {
		t.Errorf("expected 'merged' label, got %q", out)
	}
}

func TestFprint_DryRunResult(t *testing.T) {
	var buf bytes.Buffer
	results := []merge.Result{
		{Path: "secret/dst", Merged: map[string]string{"x": "y"}, DryRun: true},
	}
	merge.Fprint(&buf, results)
	if !strings.Contains(buf.String(), "dry-run") {
		t.Errorf("expected 'dry-run' label, got %q", buf.String())
	}
}

func TestFprint_ErrorResult(t *testing.T) {
	var buf bytes.Buffer
	results := []merge.Result{
		{Path: "secret/dst", Err: fmt.Errorf("boom")},
	}
	merge.Fprint(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "error") {
		t.Errorf("expected 'error' in output, got %q", out)
	}
	if !strings.Contains(out, "boom") {
		t.Errorf("expected error message in output, got %q", out)
	}
}

func TestFprint_ConflictsListed(t *testing.T) {
	var buf bytes.Buffer
	results := []merge.Result{
		{
			Path:      "secret/dst",
			Merged:    map[string]string{"a": "1"},
			Conflicts: []string{"a"},
		},
	}
	merge.Fprint(&buf, results)
	if !strings.Contains(buf.String(), "conflicts") {
		t.Errorf("expected 'conflicts' in output, got %q", buf.String())
	}
}
