// Package diff computes the difference between two sets of Vault secrets.
package diff

import (
	"sort"
)

// Op represents the type of change for a secret key.
type Op int

const (
	OpNone    Op = iota
	OpAdd
	OpRemove
	OpUpdate
)

// Change describes a single key-level difference between two secret maps.
type Change struct {
	Path     string
	Key      string
	Op       Op
	OldValue string
	NewValue string
}

// Compute returns the ordered list of changes between src and dst secret maps.
// src is the "before" state; dst is the "after" state.
func Compute(path string, src, dst map[string]string) []Change {
	var changes []Change

	// Keys present in dst
	for k, dv := range dst {
		if sv, ok := src[k]; !ok {
			changes = append(changes, Change{Path: path, Key: k, Op: OpAdd, NewValue: dv})
		} else if sv != dv {
			changes = append(changes, Change{Path: path, Key: k, Op: OpUpdate, OldValue: sv, NewValue: dv})
		} else {
			changes = append(changes, Change{Path: path, Key: k, Op: OpNone, OldValue: sv, NewValue: dv})
		}
	}

	// Keys only in src
	for k, sv := range src {
		if _, ok := dst[k]; !ok {
			changes = append(changes, Change{Path: path, Key: k, Op: OpRemove, OldValue: sv})
		}
	}

	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Key < changes[j].Key
	})
	return changes
}
