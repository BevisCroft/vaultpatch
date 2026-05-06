// Package mask provides configurable value-masking for Vault secrets.
//
// It is used throughout vaultpatch whenever secret data is displayed to the
// terminal or written to export files, ensuring that sensitive values such as
// passwords, tokens, and private keys are replaced with a safe placeholder.
//
// # Basic usage
//
//	// Use the built-in rules (covers password, token, api_key, secret, …)
//	m := mask.New(mask.DefaultRules())
//	safe := m.Apply("app/prod", rawSecrets)
//
// # Custom rules
//
//	rules := []mask.Rule{
//		{
//			PathPattern: "pki/*",
//			KeyPattern:  regexp.MustCompile(`(?i)cert|key`),
//			Replacement: "<CERT_HIDDEN>",
//		},
//	}
//	m := mask.New(rules)
//
// Apply never mutates the original map; it always returns a fresh copy.
package mask
