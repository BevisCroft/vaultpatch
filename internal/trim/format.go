package trim

import (
	"fmt"
	"io"
	"sort"
)

// Fprint writes a human-readable summary of trim results to w.
func Fprint(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "trim: no paths processed")
		return
	}

	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(w, "  ✗ %s: %v\n", r.Path, r.Err)
			continue
		}
		if len(r.Removed) == 0 {
			fmt.Fprintf(w, "  ~ %s: no matching keys\n", r.Path)
			continue
		}

		label := "trimmed"
		if r.DryRun {
			label = "would trim"
		}

		sorted := make([]string, len(r.Removed))
		copy(sorted, r.Removed)
		sort.Strings(sorted)

		for _, k := range sorted {
			fmt.Fprintf(w, "  ✓ %s: %s %q\n", r.Path, label, k)
		}
	}
}
