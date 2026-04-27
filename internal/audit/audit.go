// Package audit provides structured logging of vault operations
// performed by vaultpatch, enabling traceability of diffs and patches.
package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// OpType represents the kind of vault operation being audited.
type OpType string

const (
	OpDiff    OpType = "diff"
	OpPatch   OpType = "patch"
	OpPromote OpType = "promote"
)

// Entry is a single audit log record.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Operation OpType    `json:"operation"`
	Path      string    `json:"path"`
	DryRun    bool      `json:"dry_run"`
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
}

// Logger writes structured audit entries to an io.Writer.
type Logger struct {
	w       io.Writer
	dryRun  bool
}

// New creates a new audit Logger that writes JSON lines to w.
func New(w io.Writer, dryRun bool) *Logger {
	return &Logger{w: w, dryRun: dryRun}
}

// Record writes a single audit entry for the given operation and path.
// success indicates whether the operation completed without error.
// msg is an optional human-readable detail (e.g. error text).
func (l *Logger) Record(op OpType, path string, success bool, msg string) error {
	e := Entry{
		Timestamp: time.Now().UTC(),
		Operation: op,
		Path:      path,
		DryRun:    l.dryRun,
		Success:   success,
		Message:   msg,
	}
	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("audit: marshal entry: %w", err)
	}
	_, err = fmt.Fprintf(l.w, "%s\n", b)
	return err
}
