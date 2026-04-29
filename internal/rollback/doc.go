// Package rollback implements secret restoration from a previously saved
// snapshot. It is the inverse operation of patch/apply: given a snapshot
// file produced by the snapshot package, rollback writes each captured
// secret back to Vault, effectively undoing any changes made since the
// snapshot was taken.
//
// Typical usage:
//
//	snap, err := snapshot.Load("before.json")
//	if err != nil { ... }
//
//	applier := rollback.New(vaultClient, auditor, dryRun)
//	results, err := applier.Apply(ctx, snap)
//
// Every write is recorded via the audit.Auditor when one is provided.
// When dryRun is true the applier iterates all paths and emits audit
// records but performs no actual Vault writes.
package rollback
