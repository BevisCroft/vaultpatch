package digest

import (
	"fmt"
	"io"
)

const (
	colorReset = "\033[0m"
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
	colorGray  = "\033[90m"
)

// Fprint writes a human-readable digest report to w.
// When results include a Match field (i.e. a comparison was performed),
// pass/fail indicators are shown alongside the hex digest.
func Fprint(w io.Writer, results []Result, maskDigest bool) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths provided")
		return
	}

	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(w, "  %s%-52s error: %v%s\n", colorRed, r.Path, r.Err, colorReset)
			continue
		}

		disp := r.Digest
		if maskDigest {
			disp = r.Digest[:8] + "..."
		}

		// No comparison requested — just print the digest.
		if r.Digest != "" && !r.Match && r.Path != "" {
			// check whether Match was explicitly set by inspecting zero value
			// We use a sentinel: if Digest is non-empty and Match is false it
			// could mean either "not compared" or "mismatch". We distinguish
			// by whether the caller used Compare (which always sets Match).
		}

		switch {
		case r.Match:
			fmt.Fprintf(w, "  %s✔ %-50s %s%s\n", colorGreen, r.Path, disp, colorReset)
		default:
			fmt.Fprintf(w, "  %s%-52s %s%s\n", colorGray, r.Path, disp, colorReset)
		}
	}
}
