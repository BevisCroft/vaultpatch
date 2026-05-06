package copy

import (
	"fmt"
	"io"
)

// Fprint writes a human-readable summary of copy results to w.
func Fprint(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no copy operations to report")
		return
	}

	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(w, "  ERROR  %s → %s: %v\n", r.Src, r.Dst, r.Err)
		case r.DryRun:
			fmt.Fprintf(w, "  DRY-RUN  %s → %s\n", r.Src, r.Dst)
		default:
			fmt.Fprintf(w, "  COPIED  %s → %s\n", r.Src, r.Dst)
		}
	}
}
