package expire

import (
	"fmt"
	"io"
)

const (
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorReset  = "\033[0m"
)

// Fprint writes a human-readable expiry report to w.
func Fprint(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no expiry metadata found for any path")
		return
	}
	for _, r := range results {
		var status string
		switch {
		case r.Expired:
			status = colorRed + "EXPIRED" + colorReset
		case r.DaysLeft <= 7:
			status = colorYellow + fmt.Sprintf("expires in %d day(s)", r.DaysLeft) + colorReset
		default:
			status = colorGreen + fmt.Sprintf("ok (%d days left)", r.DaysLeft) + colorReset
		}
		line := fmt.Sprintf("  %-40s %s", r.Path, status)
		if r.Note != "" {
			line += fmt.Sprintf(" [%s]", r.Note)
		}
		fmt.Fprintln(w, line)
	}
}
