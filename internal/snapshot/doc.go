// Package snapshot provides utilities for capturing and persisting a
// point-in-time view of HashiCorp Vault KV secrets.
//
// A Snapshot records all secret paths and their key/value data beneath
// a given base path. Snapshots can be saved to disk as JSON and loaded
// back later, enabling diff and audit workflows across environments or
// deployments.
//
// Typical usage:
//
//	snap, err := snapshot.Capture(ctx, vaultClient, "secret", "myapp")
//	if err != nil { ... }
//
//	if err := snapshot.Save(snap, "before.json"); err != nil { ... }
//
//	// Later, load and compare:
//	old, err := snapshot.Load("before.json")
package snapshot
