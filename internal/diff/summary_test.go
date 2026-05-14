package diff

import (
	"bytes"
	"testing"
)

func TestSummarize_Empty(t *testing.T) {
	s := Summarize(nil)
	if s.Total != 0 || s.Added != 0 || s.Removed != 0 || s.Updated != 0 {
		t.Fatalf("expected zero summary, got %+v", s)
	}
}

func TestSummarize_AllOps(t *testing.T) {
	changes := []Change{
		{Op: OpAdd, Key: "a"},
		{Op: OpAdd, Key: "b"},
		{Op: OpRemove, Key: "c"},
		{Op: OpUpdate, Key: "d"},
		{Op: OpNone, Key: "e"},
	}
	s := Summarize(changes)
	if s.Added != 2 {
		t.Errorf("expected Added=2, got %d", s.Added)
	}
	if s.Removed != 1 {
		t.Errorf("expected Removed=1, got %d", s.Removed)
	}
	if s.Updated != 1 {
		t.Errorf("expected Updated=1, got %d", s.Updated)
	}
	if s.Total != 4 {
		t.Errorf("expected Total=4, got %d", s.Total)
	}
}

func TestFprintSummary_NoChanges(t *testing.T) {
	var buf bytes.Buffer
	FprintSummary(&buf, Summary{})
	if got := buf.String(); got != "no changes\n" {
		t.Errorf("unexpected output: %q", got)
	}
}

func TestFprintSummary_AddedOnly(t *testing.T) {
	var buf bytes.Buffer
	FprintSummary(&buf, Summary{Added: 3, Total: 3})
	got := buf.String()
	if got != "changes: +3 added (3 total)\n" {
		t.Errorf("unexpected output: %q", got)
	}
}

func TestFprintSummary_Mixed(t *testing.T) {
	var buf bytes.Buffer
	FprintSummary(&buf, Summary{Added: 1, Removed: 2, Updated: 3, Total: 6})
	got := buf.String()
	expected := "changes: +1 added, -2 removed, ~3 updated (6 total)\n"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFprintSummary_RemovedAndUpdated(t *testing.T) {
	var buf bytes.Buffer
	FprintSummary(&buf, Summary{Removed: 1, Updated: 1, Total: 2})
	got := buf.String()
	expected := "changes: -1 removed, ~1 updated (2 total)\n"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
