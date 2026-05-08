// Package summarize aggregates diff.Change slices into a concise Report that
// counts additions, updates, and removals per secret path.
//
// Typical usage:
//
//	s := summarize.New()
//	report := s.Build(changes)
//	summarize.Fprint(os.Stdout, report)
//
// The Report is intentionally decoupled from rendering so callers can
// consume the structured data (e.g. for JSON export or notifications)
// without going through the text formatter.
package summarize
