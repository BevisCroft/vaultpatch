package rotate_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/rotate"
)

func TestFprint_NoResults(t *testing.T) {
	var buf bytes.Buffer
	rotate.Fprint(&buf, nil)
	if !strings.Contains(buf.String(), "no paths rotated") {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}

func TestFprint_SuccessResult(t *testing.T) {
	var buf bytes.Buffer
	rotate.Fprint(&buf, []rotate.Result{
		{Path: "secret/app", NewKeys: []string{"password", "api_key"}, DryRun: false},
	})
	out := buf.String()
	if !strings.Contains(out, "secret/app") {
		t.Errorf("expected path in output, got: %s", out)
	}
	if !strings.Contains(out, "api_key") {
		t.Errorf("expected key name in output, got: %s", out)
	}
	if strings.Contains(out, "dry-run") {
		t.Errorf("unexpected dry-run tag for live result")
	}
}

func TestFprint_DryRunResult(t *testing.T) {
	var buf bytes.Buffer
	rotate.Fprint(&buf, []rotate.Result{
		{Path: "secret/db", NewKeys: []string{"pass"}, DryRun: true},
	})
	if !strings.Contains(buf.String(), "dry-run") {
		t.Errorf("expected dry-run tag, got: %s", buf.String())
	}
}

func TestFprint_ErrorResult(t *testing.T) {
	var buf bytes.Buffer
	rotate.Fprint(&buf, []rotate.Result{
		{Path: "secret/broken", Err: errors.New("permission denied")},
	})
	out := buf.String()
	if !strings.Contains(out, "permission denied") {
		t.Errorf("expected error message, got: %s", out)
	}
	if !strings.Contains(out, "✗") {
		t.Errorf("expected failure indicator, got: %s", out)
	}
}
