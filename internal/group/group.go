// Package group provides functionality for grouping Vault secret paths
// by a common prefix or pattern, producing a structured summary.
package group

import (
	"sort"
	"strings"
)

// Result holds the grouped secret paths under a common prefix.
type Result struct {
	Prefix string
	Paths  []string
}

// Options controls how grouping is performed.
type Options struct {
	// Depth is the number of path segments used to form the group prefix.
	// Defaults to 1 if zero.
	Depth int
}

// Grouper groups secret paths by a prefix derived from their segments.
type Grouper struct {
	opts Options
}

// New returns a new Grouper with the given options.
func New(opts Options) *Grouper {
	if opts.Depth <= 0 {
		opts.Depth = 1
	}
	return &Grouper{opts: opts}
}

// Apply groups the provided paths and returns a sorted slice of Results.
func (g *Grouper) Apply(paths []string) []Result {
	index := make(map[string][]string)

	for _, p := range paths {
		prefix := prefixAt(p, g.opts.Depth)
		index[prefix] = append(index[prefix], p)
	}

	results := make([]Result, 0, len(index))
	for prefix, ps := range index {
		sorted := make([]string, len(ps))
		copy(sorted, ps)
		sort.Strings(sorted)
		results = append(results, Result{Prefix: prefix, Paths: sorted})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Prefix < results[j].Prefix
	})

	return results
}

// prefixAt returns the first n segments of path joined by "/".
func prefixAt(path string, n int) string {
	path = strings.Trim(path, "/")
	parts := strings.SplitN(path, "/", n+1)
	if len(parts) <= n {
		return path
	}
	return strings.Join(parts[:n], "/")
}
