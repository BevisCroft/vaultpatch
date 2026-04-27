// Package promote provides functionality for promoting (copying) Vault secrets
// from a source environment to a destination environment.
//
// A typical promotion workflow:
//
//  1. Create a [Promoter] with source and destination [vault.Client] instances.
//  2. Call [Promoter.Promote] with the secret path prefix to copy.
//  3. Inspect the returned [Result] slice for per-path diffs and errors.
//
// Example:
//
//	p := promote.New(srcClient, dstClient, promote.Options{DryRun: true})
//	results, err := p.Promote(ctx, "myapp")
//	for _, r := range results {
//		fmt.Println(r.Path, r.Changes)
//	}
//
// When DryRun is true no writes are performed; the returned [Result] values
// still contain the computed [diff.Change] slice so callers can preview what
// would change.
//
// When Overwrite is false, keys that already exist in the destination are
// preserved and only new keys from the source are added.
package promote
