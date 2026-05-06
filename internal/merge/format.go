package merge

import (
	"fmt"
	"io"
	"sort"
)

// Fprint writes a human-readable summary of merge results to w.
func Fprint(w io.Writer, results []Result) {
	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(w, "  ✗ %s  error: %v\n", r.Path, r.Err)
			continue
		}

		label := "merged"
		if r.DryRun {
			label = "dry-run"
		}

		fmt.Fprintf(w, "  ✔ %s  [%s]  keys=%d", r.Path, label, len(r.Merged))

		if len(r.Conflicts) > 0 {
			sorted := make([]string, len(r.Conflicts))
			copy(sorted, r.Conflicts)
			sort.Strings(sorted)
			fmt.Fprintf(w, "  conflicts=%v", sorted)
		}

		fmt.Fprintln(w)
	}
}
