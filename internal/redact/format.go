package redact

import (
	"fmt"
	"io"
	"sort"
)

// Fprint writes a human-readable summary of which keys were redacted to w.
// original and redacted are the before/after maps from Apply.
func Fprint(w io.Writer, original, redacted map[string]string) {
	keys := make([]string, 0, len(original))
	for k := range original {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	count := 0
	for _, k := range keys {
		if original[k] != redacted[k] {
			count++
		}
	}

	if count == 0 {
		fmt.Fprintln(w, "redact: no keys matched redaction rules")
		return
	}

	fmt.Fprintf(w, "redact: %d key(s) redacted\n", count)
	for _, k := range keys {
		if original[k] != redacted[k] {
			fmt.Fprintf(w, "  - %s\n", k)
		}
	}
}
