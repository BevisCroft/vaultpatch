package search

import (
	"fmt"
	"io"
	"strings"
)

const maskPlaceholder = "***"

// Fprint writes search results to w in a human-readable format.
// When maskSecrets is true, values are replaced with ***.
func Fprint(w io.Writer, results []Result, maskSecrets bool) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no matches found")
		return
	}

	// Group by path for cleaner output.
	byPath := make(map[string][]Result)
	order := []string{}
	for _, r := range results {
		if _, seen := byPath[r.Path]; !seen {
			order = append(order, r.Path)
		}
		byPath[r.Path] = append(byPath[r.Path], r)
	}

	for _, path := range order {
		fmt.Fprintf(w, "%s\n", path)
		fmt.Fprintf(w, "%s\n", strings.Repeat("-", len(path)))
		for _, r := range byPath[path] {
			val := r.Value
			if maskSecrets {
				val = maskPlaceholder
			}
			fmt.Fprintf(w, "  %-30s %s\n", r.Key, val)
		}
		fmt.Fprintln(w)
	}
}
