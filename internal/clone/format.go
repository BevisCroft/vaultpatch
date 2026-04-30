package clone

import (
	"fmt"
	"io"
)

const (
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
)

// Fprint writes a human-readable summary of clone results to w.
func Fprint(w io.Writer, results []Result, dryRun bool) {
	if dryRun {
		fmt.Fprintln(w, colorYellow+"[dry-run] no changes written"+colorReset)
	}

	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(w, "%s✗ %s → %s: %v%s\n", colorRed, r.Src, r.Dst, r.Err, colorReset)
			continue
		}
		action := "cloned"
		if dryRun {
			action = "would clone"
		}
		fmt.Fprintf(w, "%s✓ %s %s → %s (%d keys)%s\n",
			colorGreen, action, r.Src, r.Dst, r.Keys, colorReset)
	}
}
