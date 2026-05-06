// Package verify provides functionality to assert that live Vault secret
// values match a caller-supplied set of expected key/value pairs.
//
// Usage:
//
//	v := verify.New(client, "secret")
//	results, err := v.Check(ctx, map[string]string{
//		"myapp/config/db_host": "localhost",
//		"myapp/config/db_port": "5432",
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	verify.Fprint(os.Stdout, results, false)
//
// Each key in the expected map must be of the form "path/key", where path is
// the KV path within the mount and key is the field name inside that secret.
package verify
