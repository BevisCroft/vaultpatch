package cleanup

import (
	"fmt"
	"io"
)

// Fprint writes a human-readable summary of cleanup results to w.
func Fprint(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "cleanup: no stale secrets found")
		return
	}

	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(w, "  ERROR  %s: %v\n", r.Path, r.Err)
		case r.DryRun:
			fmt.Fprintf(w, "  DRY-RUN  %s  (age: %s)\n", r.Path, r.Age.Round(1e9))
		case r.Deleted:
			fmt.Fprintf(w, "  DELETED  %s  (age: %s)\n", r.Path, r.Age.Round(1e9))
		default:
			fmt.Fprintf(w, "  SKIPPED  %s\n", r.Path)
		}
	}
}
