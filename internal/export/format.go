package export

import (
	"fmt"
	"io"
	"sort"
)

// Fprint writes a human-readable summary of an export operation to w.
func Fprint(w io.Writer, path string, format Format, count int, dryRun bool) {
	action := "Exported"
	if dryRun {
		action = "Would export"
	}
	fmt.Fprintf(w, "%s %d secret(s) as %s", action, count, format)
	if path != "-" {
		fmt.Fprintf(w, " → %s", path)
	}
	fmt.Fprintln(w)
}

// FprintKeys writes a sorted list of exported keys to w.
func FprintKeys(w io.Writer, secrets map[string]string) {
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(w, "  • %s\n", k)
	}
}
