package compare

import (
	"fmt"
	"io"
)

const (
	colorReset  = "\033[0m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
)

// Fprint writes a human-readable comparison report to w.
// When noColor is true ANSI codes are omitted.
func Fprint(w io.Writer, r Result, srcName, dstName string, noColor bool) {
	color := func(c, s string) string {
		if noColor {
			return s
		}
		return c + s + colorReset
	}

	fmt.Fprintf(w, "path: %s\n", r.Path)

	if len(r.Differ) == 0 && len(r.OnlyIn[srcName]) == 0 && len(r.OnlyIn[dstName]) == 0 {
		fmt.Fprintln(w, color(colorGreen, "  no differences"))
		return
	}

	for _, k := range r.Differ {
		fmt.Fprintf(w, "  %s  key %q differs between %s and %s\n",
			color(colorYellow, "~"), k, srcName, dstName)
	}
	for _, k := range r.OnlyIn[srcName] {
		fmt.Fprintf(w, "  %s  key %q only in %s\n",
			color(colorGreen, "+"), k, srcName)
	}
	for _, k := range r.OnlyIn[dstName] {
		fmt.Fprintf(w, "  %s  key %q only in %s\n",
			color(colorRed, "-"), k, dstName)
	}
}
