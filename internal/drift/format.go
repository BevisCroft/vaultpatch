package drift

import (
	"fmt"
	"io"

	"github.com/example/vaultpatch/internal/diff"
)

// Fprint writes a human-readable drift summary to w.
// If there are no reports, a clean message is printed.
func Fprint(w io.Writer, reports []Report, maskSecrets bool) {
	if len(reports) == 0 {
		fmt.Fprintln(w, "✓ No drift detected — live secrets match snapshot.")
		return
	}

	fmt.Fprintf(w, "⚠  Drift detected in %d path(s):\n\n", len(reports))
	for _, r := range reports {
		fmt.Fprintf(w, "  Path: %s\n", r.Path)
		diff.Fprint(w, r.Changes, maskSecrets)
		fmt.Fprintln(w)
	}
}
