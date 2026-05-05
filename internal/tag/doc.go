// Package tag provides functionality for applying and reading metadata tags
// on HashiCorp Vault secret paths.
//
// Tags are stored as Vault KV v2 custom_metadata fields, enabling teams to
// annotate secrets with environment labels, owner information, or any
// arbitrary key/value pairs without modifying the secret data itself.
//
// Usage:
//
//	manager := tag.New(vaultClient, "secret", dryRun)
//	results := manager.Apply(ctx, []string{"app/config"}, map[string]string{
//		"env":   "production",
//		"owner": "platform-team",
//	})
//	tag.Fprint(os.Stdout, results)
package tag
