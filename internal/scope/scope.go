// Package scope provides filtering of Vault secret paths by prefix, glob,
// or explicit path list, allowing commands to operate on a bounded subset
// of the secret tree.
package scope

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// Filter holds the compiled scope rules.
type Filter struct {
	prefixes []string
	globs    []string
	exact    map[string]struct{}
}

// Config describes the raw scope configuration supplied by the caller.
type Config struct {
	// Prefixes is a list of path prefixes; any secret whose path starts with
	// one of these values is included.
	Prefixes []string

	// Globs is a list of glob patterns evaluated with path.Match.
	Globs []string

	// Paths is a list of exact secret paths that must be included.
	Paths []string
}

// New builds a Filter from cfg.  An empty Config matches every path.
func New(cfg Config) (*Filter, error) {
	// validate globs at construction time so callers get early errors
	for _, g := range cfg.Globs {
		if _, err := path.Match(g, ""); err != nil {
			return nil, fmt.Errorf("scope: invalid glob %q: %w", g, err)
		}
	}
	exact := make(map[string]struct{}, len(cfg.Paths))
	for _, p := range cfg.Paths {
		exact[p] = struct{}{}
	}
	return &Filter{
		prefixes: append([]string(nil), cfg.Prefixes...),
		globs:    append([]string(nil), cfg.Globs...),
		exact:    exact,
	}, nil
}

// Match reports whether p falls within the scope.  If the filter has no
// rules at all every path is considered in-scope.
func (f *Filter) Match(p string) bool {
	if len(f.prefixes) == 0 && len(f.globs) == 0 && len(f.exact) == 0 {
		return true
	}
	if _, ok := f.exact[p]; ok {
		return true
	}
	for _, pfx := range f.prefixes {
		if strings.HasPrefix(p, pfx) {
			return true
		}
	}
	for _, g := range f.globs {
		if ok, _ := path.Match(g, p); ok {
			return true
		}
	}
	return false
}

// Apply filters paths, returning only those that match the scope.
func (f *Filter) Apply(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if f.Match(p) {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}
