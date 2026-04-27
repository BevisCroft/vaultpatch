// Package vault provides a thin wrapper around the HashiCorp Vault API client
// tailored for vaultpatch operations.
//
// It exposes a [Client] that supports reading, writing, and listing KV v2
// secrets. All methods accept a context so callers can enforce timeouts and
// cancellation.
//
// Usage:
//
//	cfg := vault.Config{
//		Address: "https://vault.example.com",
//		Token:   os.Getenv("VAULT_TOKEN"),
//		Mount:   "secret",
//	}
//	client, err := vault.NewClient(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	data, err := client.ReadSecret(ctx, "myapp/prod")
//	if err != nil {
//		log.Fatal(err)
//	}
//
Authentication is token-based. The token is expected to be supplied via the
[Config.Token] field or the VAULT_TOKEN environment variable before the client
is constructed.
package vault
