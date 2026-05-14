package group

import (
	"fmt"
	"io"
)

// Fprint writes a human-readable summary of grouped results to w.
func Fprint(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths to group")
		return
	}

	total := 0
	for _, r := range results {
		total += len(r.Paths)
	}

	fmt.Fprintf(w, "grouped %d path(s) into %d group(s)\n\n", total, len(results))

	for _, r := range results {
		fmt.Fprintf(w, "[%s] (%d)\n", r.Prefix, len(r.Paths))
		for _, p := range r.Paths {
			fmt.Fprintf(w, "  %s\n", p)
		}
	}
}
