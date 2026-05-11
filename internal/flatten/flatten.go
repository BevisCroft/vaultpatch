// Package flatten provides utilities for flattening nested Vault secret
// maps into dot-separated key paths, and for expanding them back.
package flatten

import (
	"fmt"
	"sort"
	"strings"
)

// Result holds the output of a flatten or expand operation.
type Result struct {
	// Flat maps dot-separated paths to scalar string values.
	Flat map[string]string
	// Warnings contains non-fatal issues encountered during processing.
	Warnings []string
}

// Flattener performs flatten/expand operations.
type Flattener struct {
	sep string
}

// New returns a Flattener using sep as the key separator.
// If sep is empty, "." is used.
func New(sep string) *Flattener {
	if sep == "" {
		sep = "."
	}
	return &Flattener{sep: sep}
}

// Apply flattens a nested map[string]any into a flat map[string]string.
// Nested maps are recursively descended; non-string scalar values are
// converted via fmt.Sprintf. Slices are skipped with a warning.
func (f *Flattener) Apply(input map[string]any) Result {
	out := make(map[string]string)
	var warnings []string
	f.flatten("", input, out, &warnings)
	return Result{Flat: out, Warnings: warnings}
}

func (f *Flattener) flatten(prefix string, m map[string]any, out map[string]string, warnings *[]string) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := m[k]
		fullKey := k
		if prefix != "" {
			fullKey = prefix + f.sep + k
		}
		switch val := v.(type) {
		case map[string]any:
			f.flatten(fullKey, val, out, warnings)
		case string:
			out[fullKey] = val
		case nil:
			out[fullKey] = ""
		case []any:
			*warnings = append(*warnings, fmt.Sprintf("skipping slice at key %q", fullKey))
		default:
			out[fullKey] = fmt.Sprintf("%v", val)
		}
	}
}

// Expand converts a flat map[string]string (with dot-separated keys) back
// into a nested map[string]any. Conflicts between a key used as both a
// leaf and a namespace are recorded as warnings.
func (f *Flattener) Expand(input map[string]string) Result {
	out := make(map[string]any)
	var warnings []string

	for k, v := range input {
		parts := strings.Split(k, f.sep)
		if err := setNested(out, parts, v); err != nil {
			warnings = append(warnings, fmt.Sprintf("conflict at key %q: %v", k, err))
		}
	}

	return Result{Flat: nil, Warnings: warnings}
}

func setNested(m map[string]any, parts []string, value string) error {
	if len(parts) == 1 {
		if existing, ok := m[parts[0]]; ok {
			if _, isMap := existing.(map[string]any); isMap {
				return fmt.Errorf("key already used as namespace")
			}
		}
		m[parts[0]] = value
		return nil
	}
	sub, ok := m[parts[0]]
	if !ok {
		next := make(map[string]any)
		m[parts[0]] = next
		return setNested(next, parts[1:], value)
	}
	next, ok := sub.(map[string]any)
	if !ok {
		return fmt.Errorf("key already used as leaf")
	}
	return setNested(next, parts[1:], value)
}
