// Package resolve implements inline secret reference resolution for vaultpatch.
//
// It expands placeholders of the form {{path:key}} found inside secret values
// by reading the referenced path from Vault and substituting the corresponding
// key's value. This allows secrets to reference other secrets, enabling
// dynamic composition of connection strings and configuration values without
// duplicating sensitive data across multiple paths.
//
// Example:
//
//	secrets := map[string]string{
//		"DATABASE_URL": "postgres://{{secret/db:user}}:{{secret/db:password}}@localhost/mydb",
//	}
//
//	r := resolve.New(vaultClient)
//	resolved, results := r.Apply(ctx, secrets)
//	// resolved["DATABASE_URL"] == "postgres://admin:s3cr3t@localhost/mydb"
//
// Unresolvable references (missing path, missing key, or Vault error) are left
// unchanged in the output map and reported via the returned []Result slice so
// callers can surface warnings without failing the entire operation.
package resolve
