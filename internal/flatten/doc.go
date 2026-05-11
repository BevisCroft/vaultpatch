// Package flatten provides utilities for flattening nested Vault secret maps
// into dot-separated (or custom-separated) key paths, and for expanding them
// back into nested structures.
//
// # Overview
//
// Vault KV secrets are stored as flat key/value maps, but application configs
// frequently use hierarchical naming conventions (e.g. "db.host", "db.port").
// The flatten package bridges that gap.
//
// # Usage
//
//	f := flatten.New(".")
//
//	// Flatten a nested map to dot-separated keys.
//	res := f.Apply(map[string]any{
//	    "db": map[string]any{"host": "localhost"},
//	})
//	// res.Flat => {"db.host": "localhost"}
//
//	// Expand back to nested form.
//	expanded := f.Expand(res.Flat)
//
// Slices are not supported and produce a warning in Result.Warnings.
package flatten
