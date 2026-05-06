package pin

import (
	"fmt"
	"io"
)

// Fprint writes a human-readable summary of pin/unpin results to w.
func Fprint(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths processed")
		return
	}
	for _, r := range results {
		prefix := opPrefix(r.Op)
		dryTag := ""
		if r.DryRun {
			dryTag = " [dry-run]"
		}
		if r.Err != nil {
			fmt.Fprintf(w, "  ✗ %s %s%s: %v\n", prefix, r.Path, dryTag, r.Err)
		} else {
			fmt.Fprintf(w, "  %s %s%s\n", prefix, r.Path, dryTag)
		}
	}
}

func opPrefix(op string) string {
	switch op {
	case "pin":
		return "📌"
	case "unpin":
		return "🔓"
	default:
		return "·"
	}
}
