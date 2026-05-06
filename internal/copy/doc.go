// Package copy provides functionality to duplicate Vault secrets from one
// path to another, with support for dry-run mode and bulk operations.
//
// Basic usage:
//
//	c := copy.New(vaultClient, dryRun)
//	result := c.Copy(ctx, "secret/src", "secret/dst")
//	copy.Fprint(os.Stdout, []copy.Result{result})
//
// The Copier reads the full set of key/value pairs from the source path and
// writes them verbatim to the destination path. In dry-run mode the read is
// still performed so that any access errors are surfaced, but no write is
// issued to Vault.
package copy
