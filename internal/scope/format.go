package scope

import (
	"fmt"
	"io"
	"strings"
)

// Fprint writes a human-readable summary of the active scope rules to w.
func Fprint(w io.Writer, cfg Config) {
	if len(cfg.Prefixes) == 0 && len(cfg.Globs) == 0 && len(cfg.Paths) == 0 {
		fmt.Fprintln(w, "scope: (all paths)")
		return
	}
	fmt.Fprintln(w, "scope:")
	if len(cfg.Prefixes) > 0 {
		fmt.Fprintf(w, "  prefixes : %s\n", strings.Join(cfg.Prefixes, ", "))
	}
	if len(cfg.Globs) > 0 {
		fmt.Fprintf(w, "  globs    : %s\n", strings.Join(cfg.Globs, ", "))
	}
	if len(cfg.Paths) > 0 {
		fmt.Fprintf(w, "  paths    : %s\n", strings.Join(cfg.Paths, ", "))
	}
}
