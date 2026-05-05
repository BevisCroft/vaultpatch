package tag

import (
	"fmt"
	"io"
)

// Fprint writes a human-readable summary of tag Apply results to w.
func Fprint(w io.Writer, results []Result) {
	for _, r := range results {
		prefix := "[tag]"
		if r.DryRun {
			prefix = "[tag/dry-run]"
		}
		if r.Err != nil {
			fmt.Fprintf(w, "%s ERROR %s: %v\n", prefix, r.Path, r.Err)
			continue
		}
		keys := SortedKeys(r.Tags)
		for _, k := range keys {
			fmt.Fprintf(w, "%s %s  %s=%s\n", prefix, r.Path, k, r.Tags[k])
		}
	}
}

// FprintList writes the tags for a single path to w.
func FprintList(w io.Writer, path string, tags map[string]string) {
	if len(tags) == 0 {
		fmt.Fprintf(w, "%s  (no tags)\n", path)
		return
	}
	for _, k := range SortedKeys(tags) {
		fmt.Fprintf(w, "%s  %s=%s\n", path, k, tags[k])
	}
}
