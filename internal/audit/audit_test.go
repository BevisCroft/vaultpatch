package audit_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/yourusername/vaultpatch/internal/audit"
)

func TestRecord_WritesJSONLine(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf, false)

	if err := l.Record(audit.OpPatch, "secret/data/app", true, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatal("expected non-empty output")
	}

	var e audit.Entry
	if err := json.Unmarshal([]byte(line), &e); err != nil {
		t.Fatalf("failed to unmarshal entry: %v", err)
	}

	if e.Operation != audit.OpPatch {
		t.Errorf("expected operation %q, got %q", audit.OpPatch, e.Operation)
	}
	if e.Path != "secret/data/app" {
		t.Errorf("expected path %q, got %q", "secret/data/app", e.Path)
	}
	if !e.Success {
		t.Error("expected success=true")
	}
	if e.DryRun {
		t.Error("expected dry_run=false")
	}
	if e.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestRecord_DryRunFlag(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf, true)

	_ = l.Record(audit.OpDiff, "secret/data/cfg", true, "")

	var e audit.Entry
	if err := json.Unmarshal(buf.Bytes(), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !e.DryRun {
		t.Error("expected dry_run=true")
	}
}

func TestRecord_FailureWithMessage(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf, false)

	_ = l.Record(audit.OpPromote, "secret/data/svc", false, "permission denied")

	var e audit.Entry
	if err := json.Unmarshal(buf.Bytes(), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if e.Success {
		t.Error("expected success=false")
	}
	if e.Message != "permission denied" {
		t.Errorf("expected message %q, got %q", "permission denied", e.Message)
	}
}

func TestRecord_MultipleEntries(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf, false)

	paths := []string{"secret/a", "secret/b", "secret/c"}
	for _, p := range paths {
		if err := l.Record(audit.OpPatch, p, true, ""); err != nil {
			t.Fatalf("record %q: %v", p, err)
		}
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != len(paths) {
		t.Errorf("expected %d lines, got %d", len(paths), len(lines))
	}
}
