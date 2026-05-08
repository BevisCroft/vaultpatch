package diff

import (
	"fmt"
	"io"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// Fprint writes a human-readable diff to w.
// If maskSecrets is true, values are replaced with "***".
func Fprint(w io.Writer, entries []Entry, maskSecrets bool) {
	if len(entries) == 0 {
		fmt.Fprintln(w, "No changes detected.")
		return
	}

	for _, e := range entries {
		switch e.Op {
		case OpAdded:
			fmt.Fprintf(w, "%s+ %s = %s%s\n",
				colorGreen, e.Key, maskIfNeeded(e.NewValue, maskSecrets), colorReset)
		case OpRemoved:
			fmt.Fprintf(w, "%s- %s = %s%s\n",
				colorRed, e.Key, maskIfNeeded(e.OldValue, maskSecrets), colorReset)
		case OpUpdated:
			fmt.Fprintf(w, "%s~ %s: %s → %s%s\n",
				colorYellow, e.Key,
				maskIfNeeded(e.OldValue, maskSecrets),
				maskIfNeeded(e.NewValue, maskSecrets),
				colorReset)
		}
	}

	added, removed, updated := countOps(entries)
	parts := []string{}
	if added > 0 {
		parts = append(parts, fmt.Sprintf("%d added", added))
	}
	if removed > 0 {
		parts = append(parts, fmt.Sprintf("%d removed", removed))
	}
	if updated > 0 {
		parts = append(parts, fmt.Sprintf("%d updated", updated))
	}
	fmt.Fprintf(w, "\nSummary: %s\n", strings.Join(parts, ", "))
}

func maskIfNeeded(v string, mask bool) string {
	if mask {
		return "***"
	}
	return v
}

func colorize(color, s string) string {
	return color + s + colorReset
}

func countOps(entries []Entry) (added, removed, updated int) {
	for _, e := range entries {
		switch e.Op {
		case OpAdded:
			added++
		case OpRemoved:
			removed++
		case OpUpdated:
			updated++
		}
	}
	return
}
