// Package scope provides path-level filtering for vaultpatch operations.
//
// A Filter can be constructed from three independent rule sets:
//
//   - Prefixes – include any path that starts with the given string.
//   - Globs    – include any path matched by a path.Match pattern.
//   - Paths    – include exact paths only.
//
// When all three sets are empty the filter is a no-op and every path is
// considered in-scope.  Rules are combined with OR semantics: a path is
// included if it satisfies at least one rule.
//
// Typical usage:
//
//	filter, err := scope.New(scope.Config{
//	    Prefixes: []string{"secret/prod/"},
//	})
//	if err != nil { ... }
//	matched := filter.Apply(allPaths)
package scope
