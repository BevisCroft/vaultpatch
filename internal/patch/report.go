package patch

import (
	"fmt"
	"io"

	"github.com/example/vaultpatch/internal/diff"
)

const (
	symOK   = "✓"
	symFail = "✗"
	symDry  = "~"
)

// FprintReport writes a human-readable summary of patch results to w.
func FprintReport(w io.Writer, results []Result, dryRun bool) {
	var ok, failed int

	for _, r := range results {
		sym := symOK
		if dryRun {
			sym = symDry
		} else if !r.Success {
			sym = symFail
			failed++
		} else {
			ok++
		}

		action := opLabel(r.Op)
		fmt.Fprintf(w, "  %s [%s] %s → %s\n", sym, action, r.Path, r.Key)
		if r.Err != nil {
			fmt.Fprintf(w, "      error: %v\n", r.Err)
		}
	}

	fmt.Fprintln(w)
	if dryRun {
		fmt.Fprintf(w, "Dry-run: %d change(s) would be applied.\n", len(results))
	} else {
		fmt.Fprintf(w, "Applied: %d ok, %d failed.\n", ok, failed)
	}
}

func opLabel(op diff.Op) string {
	switch op {
	case diff.OpAdd:
		return "ADD"
	case diff.OpRemove:
		return "DEL"
	case diff.OpUpdate:
		return "UPD"
	default:
		return "???"
	}
}
