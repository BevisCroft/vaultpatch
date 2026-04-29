// Package history provides apply-history tracking for vaultpatch operations.
package history

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Entry represents a single recorded apply operation.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Environment string    `json:"environment"`
	Operator    string    `json:"operator"`
	DryRun      bool      `json:"dry_run"`
	Changes     int       `json:"changes"`
	Failures    int       `json:"failures"`
	Note        string    `json:"note,omitempty"`
}

// Recorder appends history entries to a newline-delimited JSON file.
type Recorder struct {
	path string
}

// New returns a Recorder that writes to the given file path.
func New(path string) *Recorder {
	return &Recorder{path: path}
}

// Record appends e to the history file.
func (r *Recorder) Record(e Entry) error {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	f, err := os.OpenFile(r.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("history: open %s: %w", r.path, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err := enc.Encode(e); err != nil {
		return fmt.Errorf("history: encode entry: %w", err)
	}
	return nil
}

// Load reads all entries from the history file.
func Load(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("history: open %s: %w", path, err)
	}
	defer f.Close()
	var entries []Entry
	dec := json.NewDecoder(f)
	for dec.More() {
		var e Entry
		if err := dec.Decode(&e); err != nil {
			return nil, fmt.Errorf("history: decode entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}
