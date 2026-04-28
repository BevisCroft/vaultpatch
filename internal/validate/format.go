package validate

import (
	"fmt"
	"io"
)

const (
	ansiRed    = "\033[31m"
	ansiYellow = "\033[33m"
	ansiReset  = "\033[0m"
)

// Fprint writes a human-readable validation report to w.
// It prints a summary line followed by each issue.
func Fprint(w io.Writer, res *Result, noColor bool) {
	if len(res.Issues) == 0 {
		fmt.Fprintln(w, "validation passed: no issues found")
		return
	}

	errorCount, warnCount := 0, 0
	for _, iss := range res.Issues {
		if iss.Severity == "error" {
			errorCount++
		} else {
			warnCount++
		}
	}

	fmt.Fprintf(w, "validation complete: %d error(s), %d warning(s)\n", errorCount, warnCount)

	for _, iss := range res.Issues {
		line := iss.String()
		if !noColor {
			switch iss.Severity {
			case "error":
				line = ansiRed + line + ansiReset
			case "warning":
				line = ansiYellow + line + ansiReset
			}
		}
		fmt.Fprintln(w, line)
	}
}
