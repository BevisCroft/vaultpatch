// Package rotate provides secret rotation for HashiCorp Vault paths.
//
// A Rotator reads the current key-value pairs at a given Vault path,
// generates fresh values for every key using a caller-supplied Generator
// function, and writes the updated data back atomically. All operations
// are recorded via the audit package and honour a dry-run flag so that
// changes can be previewed before being committed.
//
// Typical usage:
//
//	gen := func(path, key string) (string, error) {
//	    return uuid.NewString(), nil
//	}
//	rot := rotate.New(client, client, auditor, dryRun)
//	result := rot.Apply(ctx, "secret/app/db", gen)
//	rotate.Fprint(os.Stdout, []rotate.Result{result})
package rotate
