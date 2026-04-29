// Package schedule provides functionality for planning and describing
// deferred or recurring vaultpatch operations (e.g. scheduled applies).
package schedule

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Frequency describes how often a scheduled job should run.
type Frequency string

const (
	FrequencyOnce  Frequency = "once"
	FrequencyDaily Frequency = "daily"
	FrequencyWeekly Frequency = "weekly"
)

// Entry represents a single scheduled vaultpatch operation.
type Entry struct {
	ID        string    `json:"id"`
	Operation string    `json:"operation"` // e.g. "apply", "promote", "rollback"
	Mount     string    `json:"mount"`
	Frequency Frequency `json:"frequency"`
	NextRunAt time.Time `json:"next_run_at"`
	DryRun    bool      `json:"dry_run"`
	CreatedAt time.Time `json:"created_at"`
}

// Schedule holds a collection of scheduled entries persisted to disk.
type Schedule struct {
	Entries []Entry `json:"entries"`
}

// Load reads a schedule file from path. Returns an empty Schedule if the
// file does not exist.
func Load(path string) (*Schedule, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Schedule{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("schedule: read %q: %w", path, err)
	}
	var s Schedule
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("schedule: unmarshal: %w", err)
	}
	return &s, nil
}

// Save persists the schedule to path.
func (s *Schedule) Save(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("schedule: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("schedule: write %q: %w", path, err)
	}
	return nil
}

// Add appends a new entry to the schedule.
func (s *Schedule) Add(e Entry) {
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}
	s.Entries = append(s.Entries, e)
}

// Remove deletes the entry with the given ID. Returns an error if not found.
func (s *Schedule) Remove(id string) error {
	for i, e := range s.Entries {
		if e.ID == id {
			s.Entries = append(s.Entries[:i], s.Entries[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("schedule: entry %q not found", id)
}

// Due returns all entries whose NextRunAt is at or before now.
func (s *Schedule) Due(now time.Time) []Entry {
	var due []Entry
	for _, e := range s.Entries {
		if !e.NextRunAt.After(now) {
			due = append(due, e)
		}
	}
	return due
}
