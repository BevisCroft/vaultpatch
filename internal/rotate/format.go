package rotate

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// Fprint writes a human-readable summary of rotation results to w.
func Fprint(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths rotated")
		return
	}

	for _, r := range results {
		tag := ""
		if r.DryRun {
			tag = " [dry-run]"
		}
		if r.Err != nil {
			fmt.Fprintf(w, "  ✗ %s%s — error: %v\n", r.Path, tag, r.Err)
			continue
		}
		keys := append([]string(nil), r.NewKeys...)
		sort.Strings(keys)
		fmt.Fprintf(w, "  ✓ %s%s — rotated keys: %s\n", r.Path, tag, strings.Join(keys, ", "))
	}
}
