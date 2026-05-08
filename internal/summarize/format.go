package summarize

import (
	"fmt"
	"io"
)

// Fprint writes a human-readable summary report to w.
func Fprint(w io.Writer, r Report) {
	if len(r.Paths) == 0 {
		fmt.Fprintln(w, "No changes detected.")
		return
	}

	fmt.Fprintf(w, "Summary: +%d added  ~%d updated  -%d removed\n",
		r.TotalAdded, r.TotalUpdated, r.TotalRemoved)
	fmt.Fprintln(w)

	for _, ps := range r.Paths {
		if ps.Added+ps.Removed+ps.Updated == 0 {
			continue
		}
		fmt.Fprintf(w, "  %s\n", ps.Path)
		if ps.Added > 0 {
			fmt.Fprintf(w, "    + %d added\n", ps.Added)
		}
		if ps.Updated > 0 {
			fmt.Fprintf(w, "    ~ %d updated\n", ps.Updated)
		}
		if ps.Removed > 0 {
			fmt.Fprintf(w, "    - %d removed\n", ps.Removed)
		}
	}
}
