// Package clone provides functionality for copying Vault secrets from one
// path to another, with optional dry-run support and audit logging.
//
// Usage:
//
//	cloner := clone.New(vaultClient, vaultClient, auditor, dryRun)
//	result := cloner.Clone(ctx, "secret/staging/app", "secret/prod/app")
//	clone.Fprint(os.Stdout, []clone.Result{result}, dryRun)
//
// In dry-run mode, secrets are read from the source but never written to the
// destination. All operations are recorded via the audit logger regardless of
// dry-run status.
package clone
