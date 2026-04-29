package drift_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/example/vaultpatch/internal/diff"
	"github.com/example/vaultpatch/internal/drift"
)

func TestFprint_NoDrift(t *testing.T) {
	var buf bytes.Buffer
	drift.Fprint(&buf, nil, false)
	if !strings.Contains(buf.String(), "No drift detected") {
		t.Errorf("expected clean message, got: %q", buf.String())
	}
}

func TestFprint_WithDrift(t *testing.T) {
	reports := []drift.Report{
		{
			Path: "app/db",
			Changes: []diff.Change{
				{Op: diff.OpUpdated, Key: "password", OldValue: "old", NewValue: "new"},
			},
		},
	}

	var buf bytes.Buffer
	drift.Fprint(&buf, reports, false)
	out := buf.String()

	if !strings.Contains(out, "Drift detected in 1 path") {
		t.Errorf("expected drift header, got: %q", out)
	}
	if !strings.Contains(out, "app/db") {
		t.Errorf("expected path in output, got: %q", out)
	}
}

func TestFprint_MaskSecrets(t *testing.T) {
	reports := []drift.Report{
		{
			Path: "app/creds",
			Changes: []diff.Change{
				{Op: diff.OpUpdated, Key: "secret_key", OldValue: "abc", NewValue: "xyz"},
			},
		},
	}

	var buf bytes.Buffer
	drift.Fprint(&buf, reports, true)
	out := buf.String()

	if strings.Contains(out, "abc") || strings.Contains(out, "xyz") {
		t.Errorf("expected secrets to be masked, got: %q", out)
	}
}
