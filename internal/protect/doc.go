// Package protect implements path-level write protection for Vault secrets.
//
// A protected path has a sentinel metadata key (vaultpatch/protected) written
// alongside its normal key-value data. Other vaultpatch commands that mutate
// secrets should call IsProtected before writing to honour this contract.
//
// Usage:
//
//	p := protect.New(vaultClient, dryRun)
//
//	// Mark paths as protected.
//	results := p.Protect(ctx, []string{"secret/prod/db"})
//
//	// Check before writing elsewhere.
//	guarded, err := p.IsProtected(ctx, "secret/prod/db")
//
//	// Remove protection.
//	results = p.Unprotect(ctx, []string{"secret/prod/db"})
package protect
