// Package patch provides the ability to apply a computed diff ([]diff.Entry)
// to a live HashiCorp Vault instance.
//
// Basic usage:
//
//	entries, _ := diff.Compute(src, dst)
//	applier := patch.NewApplier(vaultClient, dryRun)
//	results := applier.Apply(ctx, entries)
//	patch.FprintReport(os.Stdout, results, dryRun)
//
// When dryRun is true, no writes are made to Vault; results are still
// returned so callers can preview what would change.
//
// The Applier relies on vault.Client methods WriteSecretKey and
// DeleteSecretKey to perform individual key-level mutations, keeping
// unrelated keys in the same Vault path untouched.
package patch
