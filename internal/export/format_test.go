package export

import (
	"bytes"
	"strings"
	"testing"
)

func TestFprint_DryRun(t *testing.T) {
	var buf bytes.Buffer
	Fprint(&buf, "out.json", FormatJSON, 3, true)
	got := buf.String()
	if !strings.Contains(got, "Would export") {
		t.Errorf("expected dry-run label, got: %q", got)
	}
	if !strings.Contains(got, "3") {
		t.Errorf("expected count 3, got: %q", got)
	}
}

func TestFprint_Live(t *testing.T) {
	var buf bytes.Buffer
	Fprint(&buf, "/tmp/secrets.yaml", FormatYAML, 5, false)
	got := buf.String()
	if !strings.Contains(got, "Exported") {
		t.Errorf("expected Exported label, got: %q", got)
	}
	if !strings.Contains(got, "/tmp/secrets.yaml") {
		t.Errorf("expected path in output, got: %q", got)
	}
}

func TestFprint_Stdout(t *testing.T) {
	var buf bytes.Buffer
	Fprint(&buf, "-", FormatEnv, 2, false)
	got := buf.String()
	if strings.Contains(got, "→") {
		t.Errorf("stdout export should not show path arrow, got: %q", got)
	}
}

func TestFprintKeys(t *testing.T) {
	var buf bytes.Buffer
	secrets := map[string]string{"zebra": "z", "alpha": "a", "mid": "m"}
	FprintKeys(&buf, secrets)
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "alpha") {
		t.Errorf("expected first key to be alpha, got: %q", lines[0])
	}
	if !strings.Contains(lines[2], "zebra") {
		t.Errorf("expected last key to be zebra, got: %q", lines[2])
	}
}
