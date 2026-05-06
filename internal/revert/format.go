package revert

import (
	"fmt"
	"io"
)

// Fprint writes a human-readable summary of revert results to w.
func Fprint(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "revert: no paths to process")
		return
	}

	for _, r := range results {
		switch {
		case r.Skipped:
			fmt.Fprintf(w, "  ~ %s (not in snapshot, skipped)\n", r.Path)
		case r.Err != nil:
			fmt.Fprintf(w, "  ✗ %s — %s\n", r.Path, r.Err)
		case r.DryRun:
			fmt.Fprintf(w, "  ○ %s (dry-run)\n", r.Path)
		default:
			fmt.Fprintf(w, "  ✓ %s\n", r.Path)
		}
	}

	var errs, skipped, ok int
	for _, r := range results {
		switch {
		case r.Err != nil:
			errs++
		case r.Skipped:
			skipped++
		default:
			ok++
		}
	}

	fmt.Fprintf(w, "\nrevert: %d reverted, %d skipped, %d errors\n", ok, skipped, errs)
}
