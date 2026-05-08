// Package diff computes the difference between two sets of Vault secrets.
package diff

import (
	"sort"
)

// Op represents the type of change for a secret key.
type Op int

const (
	OpNone    Op = iota
	OpAdded      // key exists in new but not old
	OpRemoved    // key exists in old but not new
	OpUpdated    // key exists in both but value changed
)

// Entry describes a single diff entry for one key.
type Entry struct {
	Key      string
	OldValue string
	NewValue string
	Op       Op
}

// Compute returns the diff between oldSecrets and newSecrets.
// Keys present in both with identical values are omitted.
func Compute(oldSecrets, newSecrets map[string]string) []Entry {
	seen := make(map[string]bool)
	var entries []Entry

	for k, newVal := range newSecrets {
		seen[k] = true
		oldVal, exists := oldSecrets[k]
		switch {
		case !exists:
			entries = append(entries, Entry{Key: k, NewValue: newVal, Op: OpAdded})
		case oldVal != newVal:
			entries = append(entries, Entry{Key: k, OldValue: oldVal, NewValue: newVal, Op: OpUpdated})
		}
	}

	for k, oldVal := range oldSecrets {
		if !seen[k] {
			entries = append(entries, Entry{Key: k, OldValue: oldVal, Op: OpRemoved})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Key < entries[j].Key
	})

	return entries
}
