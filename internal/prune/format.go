package prune

import (
	"fmt"
	"io"
)

// Fprint writes a human-readable summary of prune results to w.
func Fprint(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "prune: no secrets evaluated")
		return
	}

	var pruned, skipped, failed int
	for _, r := range results {
		switch {
		case r.Err != nil:
			failed++
			fmt.Fprintf(w, "  ERROR   %s: %v\n", r.Path, r.Err)
		case r.Skipped:
			skipped++
			fmt.Fprintf(w, "  SKIP    %s\n", r.Path)
		case r.DryRun:
			pruned++
			fmt.Fprintf(w, "  DRY-RUN %s  (would prune)\n", r.Path)
		default:
			pruned++
			fmt.Fprintf(w, "  PRUNED  %s\n", r.Path)
		}
	}

	fmt.Fprintf(w, "\nSummary: %d pruned, %d skipped, %d failed\n", pruned, skipped, failed)
}
