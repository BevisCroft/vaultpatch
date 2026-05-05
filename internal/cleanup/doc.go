// Package cleanup provides tooling to identify and remove stale Vault secrets
// that have not been updated within a configurable time window.
//
// # Overview
//
// A Cleaner is constructed with a VaultClient, an Auditor, a maximum age
// duration, and a dry-run flag. Calling Run against a path prefix will:
//
//  1. List all secrets under the prefix.
//  2. Read each secret and inspect its "updated_at" metadata field.
//  3. Mark secrets whose age exceeds MaxAge as stale.
//  4. Delete stale secrets (unless DryRun is true).
//  5. Record every action via the Auditor.
//
// # Usage
//
//	cleaner := cleanup.New(client, auditor, 30*24*time.Hour, false)
//	results, err := cleaner.Run(ctx, "secret/prod")
//	cleanup.Fprint(os.Stdout, results)
package cleanup
