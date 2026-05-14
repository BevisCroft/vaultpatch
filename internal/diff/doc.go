// Package diff computes and formats the difference between two sets of
// Vault secret key-value pairs.
//
// # Computing a diff
//
// Use [Compute] to compare a source map against a destination map. It returns
// a slice of [Change] values, each describing a single key-level operation:
// add, remove, update, or none (unchanged).
//
// # Formatting
//
// [Fprint] renders the changes in a colourised, unified-diff-style output
// suitable for a terminal. Sensitive values are automatically masked when the
// mask flag is set.
//
// # Summarising
//
// [Summarize] aggregates a []Change into a [Summary] struct that counts adds,
// removes, and updates. [FprintSummary] writes a compact one-line description
// of the summary to any io.Writer.
package diff
