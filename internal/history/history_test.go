package history_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/your-org/vaultpatch/internal/history"
)

func tempFile(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "history.jsonl")
}

func TestRecord_WritesEntry(t *testing.T) {
	p := tempFile(t)
	r := history.New(p)
	e := history.Entry{
		Environment: "staging",
		Operator:    "alice",
		DryRun:      false,
		Changes:     3,
		Failures:    0,
	}
	if err := r.Record(e); err != nil {
		t.Fatalf("Record: %v", err)
	}
	entries, err := history.Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Environment != "staging" {
		t.Errorf("expected staging, got %s", entries[0].Environment)
	}
	if entries[0].Changes != 3 {
		t.Errorf("expected 3 changes, got %d", entries[0].Changes)
	}
}

func TestRecord_SetsTimestampIfZero(t *testing.T) {
	p := tempFile(t)
	r := history.New(p)
	before := time.Now().UTC()
	if err := r.Record(history.Entry{Environment: "prod"}); err != nil {
		t.Fatalf("Record: %v", err)
	}
	entries, _ := history.Load(p)
	if entries[0].Timestamp.Before(before) {
		t.Error("timestamp should be set automatically")
	}
}

func TestRecord_MultipleEntries(t *testing.T) {
	p := tempFile(t)
	r := history.New(p)
	for i := 0; i < 5; i++ {
		if err := r.Record(history.Entry{Changes: i}); err != nil {
			t.Fatalf("Record[%d]: %v", i, err)
		}
	}
	entries, err := history.Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(entries) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(entries))
	}
}

func TestLoad_MissingFile(t *testing.T) {
	entries, err := history.Load("/nonexistent/history.jsonl")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	if entries != nil {
		t.Error("expected nil entries for missing file")
	}
}

func TestRecord_DryRunFlagged(t *testing.T) {
	p := tempFile(t)
	r := history.New(p)
	_ = r.Record(history.Entry{DryRun: true, Environment: "dev"})
	entries, _ := history.Load(p)
	if !entries[0].DryRun {
		t.Error("expected DryRun=true")
	}
}

func TestRecord_InvalidPath(t *testing.T) {
	r := history.New("/no/such/dir/history.jsonl")
	err := r.Record(history.Entry{})
	if err == nil {
		t.Error("expected error for invalid path")
	}
	_ = os.Remove("/no/such/dir/history.jsonl")
}
