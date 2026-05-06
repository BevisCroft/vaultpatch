package archive

import (
	"fmt"
	"io"
)

// Fprint writes a human-readable summary of archive results to w.
func Fprint(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "archive: no paths specified")
		return
	}

	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(w, "  ✗ %s → %s  error: %v\n", r.Path, r.Archive, r.Err)
		case r.DryRun:
			fmt.Fprintf(w, "  ~ %s → %s  (dry-run)\n", r.Path, r.Archive)
		default:
			fmt.Fprintf(w, "  ✓ %s → %s\n", r.Path, r.Archive)
		}
	}
}
