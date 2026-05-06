package verify

import (
	"fmt"
	"io"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
)

// Fprint writes a human-readable verification report to w.
func Fprint(w io.Writer, results []Result, maskValues bool) {
	if len(results) == 0 {
		fmt.Fprintln(w, "verify: no expectations to check")
		return
	}

	passed, failed, errored := 0, 0, 0
	for _, r := range results {
		if r.Err != nil {
			errored++
			fmt.Fprintf(w, "%s[ERROR]%s  %s: %v\n", colorYellow, colorReset, r.Path, r.Err)
			continue
		}
		if r.Match {
			passed++
			fmt.Fprintf(w, "%s[PASS]%s   %s/%s\n", colorGreen, colorReset, r.Path, r.Key)
		} else {
			failed++
			actual := r.Actual
			expected := r.Expected
			if maskValues {
				actual = "***"
				expected = "***"
			}
			fmt.Fprintf(w, "%s[FAIL]%s   %s/%s: expected=%q actual=%q\n",
				colorRed, colorReset, r.Path, r.Key, expected, actual)
		}
	}
	fmt.Fprintf(w, "\nverify: %d passed, %d failed, %d errored\n", passed, failed, errored)
}
