package access

import (
	"fmt"
	"io"
)

const (
	symAllow = "✓"
	symDeny  = "✗"
)

// Fprint writes a human-readable summary of access results to w.
func Fprint(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no access checks to display")
		return
	}

	allowed := 0
	for _, r := range results {
		if r.Allowed {
			allowed++
		}
	}

	fmt.Fprintf(w, "access check results: %d/%d allowed\n", allowed, len(results))

	for _, r := range results {
		sym := symDeny
		if r.Allowed {
			sym = symAllow
		}
		fmt.Fprintf(w, "  %s  [%-6s] %s\n", sym, r.Op, r.Path)
		if !r.Allowed {
			fmt.Fprintf(w, "         reason: %s\n", r.Reason)
		}
	}
}
