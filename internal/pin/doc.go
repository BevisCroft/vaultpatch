// Package pin provides secret path pinning for vaultpatch.
//
// Pinning a secret path records metadata in Vault's KV v2 custom_metadata
// fields to signal that the path should not be overwritten by automated
// patch or promote operations until explicitly unpinned.
//
// Usage:
//
//	p := pin.New(client, "secret", false)
//	result := p.Pin(ctx, "myapp/db", "alice", "release freeze")
//	result = p.Unpin(ctx, "myapp/db")
//
// The IsPinned helper lets other modules (patch, promote, rotate) gate
// writes behind a pin check before proceeding.
package pin
