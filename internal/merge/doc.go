// Package merge provides functionality to merge Vault secrets from a source
// path into a destination path using a configurable conflict-resolution
// strategy.
//
// Three strategies are supported:
//
//   - ours   – keep the destination value when a key exists in both paths
//   - theirs – overwrite with the source value on conflict
//   - error  – abort the merge and return an error on the first conflict
//
// All write operations can be previewed with the dry-run flag, which performs
// every read but skips the final write to Vault.
package merge
