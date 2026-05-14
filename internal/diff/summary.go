package diff

import (
	"fmt"
	"io"
	"strings"
)

// Summary holds aggregated statistics for a diff result.
type Summary struct {
	Added   int
	Removed int
	Updated int
	Total   int
}

// Summarize computes a Summary from a slice of Changes.
func Summarize(changes []Change) Summary {
	var s Summary
	for _, c := range changes {
		switch c.Op {
		case OpAdd:
			s.Added++
		case OpRemove:
			s.Removed++
		case OpUpdate:
			s.Updated++
		}
	}
	s.Total = s.Added + s.Removed + s.Updated
	return s
}

// FprintSummary writes a human-readable one-line summary to w.
func FprintSummary(w io.Writer, s Summary) {
	if s.Total == 0 {
		fmt.Fprintln(w, "no changes")
		return
	}
	parts := make([]string, 0, 3)
	if s.Added > 0 {
		parts = append(parts, fmt.Sprintf("+%d added", s.Added))
	}
	if s.Removed > 0 {
		parts = append(parts, fmt.Sprintf("-%d removed", s.Removed))
	}
	if s.Updated > 0 {
		parts = append(parts, fmt.Sprintf("~%d updated", s.Updated))
	}
	fmt.Fprintf(w, "changes: %s (%d total)\n", strings.Join(parts, ", "), s.Total)
}
