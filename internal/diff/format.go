package diff

import (
	"fmt"
	"io"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// FormatOptions controls how a diff result is rendered.
type FormatOptions struct {
	Color   bool
	MaskValues bool
}

// Fprint writes a human-readable diff to w using the provided options.
func Fprint(w io.Writer, result *Result, opts FormatOptions) {
	if !result.HasChanges() {
		fmt.Fprintln(w, "No changes detected.")
		return
	}

	for _, c := range result.Changes {
		switch c.Type {
		case ChangeAdded:
			line := fmt.Sprintf("+ %s = %s", c.Key, maskIfNeeded(c.NewValue, opts.MaskValues))
			fmt.Fprintln(w, colorize(line, colorGreen, opts.Color))
		case ChangeRemoved:
			line := fmt.Sprintf("- %s = %s", c.Key, maskIfNeeded(c.OldValue, opts.MaskValues))
			fmt.Fprintln(w, colorize(line, colorRed, opts.Color))
		case ChangeUpdated:
			oldVal := maskIfNeeded(c.OldValue, opts.MaskValues)
			newVal := maskIfNeeded(c.NewValue, opts.MaskValues)
			line := fmt.Sprintf("~ %s: %s -> %s", c.Key, oldVal, newVal)
			fmt.Fprintln(w, colorize(line, colorYellow, opts.Color))
		}
	}
}

func colorize(s, color string, enabled bool) string {
	if !enabled {
		return s
	}
	return color + s + colorReset
}

func maskIfNeeded(val string, mask bool) string {
	if !mask || val == "" {
		return val
	}
	return strings.Repeat("*", 8)
}
