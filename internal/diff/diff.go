// Package diff provides utilities for computing differences between
// Vault secret maps across environments.
package diff

import "sort"

// ChangeType represents the type of change detected for a secret key.
type ChangeType string

const (
	ChangeAdded   ChangeType = "added"
	ChangeRemoved ChangeType = "removed"
	ChangeUpdated ChangeType = "updated"
)

// Change represents a single key-level change between two secret maps.
type Change struct {
	Key      string
	Type     ChangeType
	OldValue string
	NewValue string
}

// Result holds the full diff result between a source and target secret map.
type Result struct {
	Changes []Change
}

// HasChanges returns true if the diff result contains any changes.
func (r *Result) HasChanges() bool {
	return len(r.Changes) > 0
}

// Compute calculates the difference between src and dst secret maps.
// src is the baseline (e.g. current environment), dst is the desired state.
func Compute(src, dst map[string]string) *Result {
	result := &Result{}

	for key, dstVal := range dst {
		if srcVal, exists := src[key]; !exists {
			result.Changes = append(result.Changes, Change{
				Key:      key,
				Type:     ChangeAdded,
				NewValue: dstVal,
			})
		} else if srcVal != dstVal {
			result.Changes = append(result.Changes, Change{
				Key:      key,
				Type:     ChangeUpdated,
				OldValue: srcVal,
				NewValue: dstVal,
			})
		}
	}

	for key, srcVal := range src {
		if _, exists := dst[key]; !exists {
			result.Changes = append(result.Changes, Change{
				Key:      key,
				Type:     ChangeRemoved,
				OldValue: srcVal,
			})
		}
	}

	sort.Slice(result.Changes, func(i, j int) bool {
		return result.Changes[i].Key < result.Changes[j].Key
	})

	return result
}
