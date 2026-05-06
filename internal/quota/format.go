package quota

import (
	"fmt"
	"io"
)

// Fprint writes a human-readable quota report to w.
func Fprint(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no quota rules defined")
		return
	}
	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(w, "  ERROR  %s: %v\n", r.Path, r.Err)
			continue
		}
		status := "OK     "
		if r.Exceeds {
			status = "EXCEEDS"
		}
		fmt.Fprintf(w, "  %s  %s  (%d/%d)\n", status, r.Path, r.Current, r.Limit)
	}
}
