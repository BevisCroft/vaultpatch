// Package digest provides SHA-256 fingerprinting for HashiCorp Vault secret
// paths.
//
// # Overview
//
// A Digester reads a secret path from Vault and produces a deterministic
// hex-encoded SHA-256 digest over its key/value pairs. Keys are sorted before
// hashing so the result is stable regardless of map iteration order.
//
// # Usage
//
//	d := digest.New(vaultClient)
//
//	// Compute a fresh digest:
//	result := d.Compute("secret/data/myapp/prod")
//
//	// Compare against a previously stored baseline:
//	result = d.Compare("secret/data/myapp/prod", storedDigest)
//	if !result.Match {
//	    log.Println("secret has changed!")
//	}
package digest
