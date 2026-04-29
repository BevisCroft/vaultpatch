package schedule_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/vaultpatch/internal/schedule"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "schedule.json")
}

func TestLoad_MissingFile(t *testing.T) {
	s, err := schedule.Load("/nonexistent/schedule.json")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if len(s.Entries) != 0 {
		t.Fatalf("expected empty schedule, got %d entries", len(s.Entries))
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	path := tempPath(t)
	s := &schedule.Schedule{}
	s.Add(schedule.Entry{
		ID:        "entry-1",
		Operation: "apply",
		Mount:     "secret",
		Frequency: schedule.FrequencyDaily,
		NextRunAt: time.Now().UTC().Add(24 * time.Hour),
		DryRun:    false,
	})

	if err := s.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := schedule.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(loaded.Entries))
	}
	if loaded.Entries[0].ID != "entry-1" {
		t.Errorf("expected ID entry-1, got %s", loaded.Entries[0].ID)
	}
}

func TestAdd_SetsCreatedAt(t *testing.T) {
	s := &schedule.Schedule{}
	before := time.Now().UTC()
	s.Add(schedule.Entry{ID: "x", Operation: "promote"})
	if s.Entries[0].CreatedAt.Before(before) {
		t.Error("expected CreatedAt to be set to approximately now")
	}
}

func TestRemove_ExistingEntry(t *testing.T) {
	s := &schedule.Schedule{}
	s.Add(schedule.Entry{ID: "del-me"})
	if err := s.Remove("del-me"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if len(s.Entries) != 0 {
		t.Errorf("expected 0 entries after removal, got %d", len(s.Entries))
	}
}

func TestRemove_NotFound(t *testing.T) {
	s := &schedule.Schedule{}
	if err := s.Remove("ghost"); err == nil {
		t.Error("expected error when removing non-existent entry")
	}
}

func TestDue_ReturnsOverdueEntries(t *testing.T) {
	now := time.Now().UTC()
	s := &schedule.Schedule{}
	s.Add(schedule.Entry{ID: "past", NextRunAt: now.Add(-1 * time.Hour)})
	s.Add(schedule.Entry{ID: "future", NextRunAt: now.Add(1 * time.Hour)})

	due := s.Due(now)
	if len(due) != 1 {
		t.Fatalf("expected 1 due entry, got %d", len(due))
	}
	if due[0].ID != "past" {
		t.Errorf("expected due entry ID 'past', got %s", due[0].ID)
	}
}

func TestSave_FilePermissions(t *testing.T) {
	path := tempPath(t)
	s := &schedule.Schedule{}
	if err := s.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("expected file mode 0600, got %o", info.Mode().Perm())
	}
}
