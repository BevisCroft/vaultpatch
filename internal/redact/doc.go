// Package redact implements configurable secret-key redaction for vaultpatch.
//
// It allows callers to define patterns that match sensitive Vault secret keys
// (e.g. "password", "token") and replace their values with a placeholder
// before the data is printed, exported, or compared.
//
// # Basic usage
//
//	r := redact.New(redact.DefaultRules())
//	safe := r.Apply(rawSecrets)
//	redact.Fprint(os.Stdout, rawSecrets, safe)
//
// Custom rules can be provided to extend or replace the defaults:
//
//	rules := []redact.Rule{
//		{Pattern: "ssn", Replacement: "[SSN REDACTED]"},
//	}
//	r := redact.New(rules)
package redact
