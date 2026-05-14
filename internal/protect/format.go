package protect

import (
	"fmt"
	"io"
)

// Fprint writes a human-readable summary of protect/unprotect results to w.
func Fprint(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths specified")
		return
	}
	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(w, "  ✗ %s: %v\n", r.Path, r.Err)
		case r.DryRun:
			fmt.Fprintf(w, "  ~ %s (%s) [dry-run]\n", r.Path, actionLabel(r.Protected))
		default:
			fmt.Fprintf(w, "  ✓ %s (%s)\n", r.Path, actionLabel(r.Protected))
		}
	}
	ok, failed := countResults(results)
	fmt.Fprintf(w, "\n%d succeeded, %d failed\n", ok, failed)
}

func actionLabel(protect bool) string {
	if protect {
		return "protected"
	}
	return "unprotected"
}

func countResults(results []Result) (ok, failed int) {
	for _, r := range results {
		if r.Err != nil {
			failed++
		} else {
			ok++
		}
	}
	return
}
