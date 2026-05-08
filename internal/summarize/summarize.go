// Package summarize produces a high-level summary of diff changes across
// multiple Vault secret paths, suitable for reports and notifications.
package summarize

import (
	"sort"

	"github.com/example/vaultpatch/internal/diff"
)

// PathSummary holds aggregated change counts for a single path.
type PathSummary struct {
	Path    string
	Added   int
	Removed int
	Updated int
	Unchanged int
}

// Report is the top-level summary across all paths.
type Report struct {
	Paths      []PathSummary
	TotalAdded   int
	TotalRemoved int
	TotalUpdated int
}

// Summarizer aggregates diff.Change slices into a Report.
type Summarizer struct{}

// New returns a new Summarizer.
func New() *Summarizer { return &Summarizer{} }

// Build groups the provided changes by path and returns a Report.
func (s *Summarizer) Build(changes []diff.Change) Report {
	index := map[string]*PathSummary{}

	for _, c := range changes {
		ps, ok := index[c.Path]
		if !ok {
			ps = &PathSummary{Path: c.Path}
			index[c.Path] = ps
		}
		switch c.Op {
		case diff.OpAdd:
			ps.Added++
		case diff.OpRemove:
			ps.Removed++
		case diff.OpUpdate:
			ps.Updated++
		default:
			ps.Unchanged++
		}
	}

	var report Report
	for _, ps := range index {
		report.Paths = append(report.Paths, *ps)
		report.TotalAdded += ps.Added
		report.TotalRemoved += ps.Removed
		report.TotalUpdated += ps.Updated
	}
	sort.Slice(report.Paths, func(i, j int) bool {
		return report.Paths[i].Path < report.Paths[j].Path
	})
	return report
}
