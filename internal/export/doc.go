// Package export serializes Vault secret maps to disk or stdout in
// multiple formats: JSON, YAML, and dotenv (.env).
//
// # Usage
//
//	exporter, err := export.New("json")
//	if err != nil {
//		log.Fatal(err)
//	}
//	if err := exporter.Write(secrets, "/tmp/secrets.json"); err != nil {
//		log.Fatal(err)
//	}
//
// Pass "-" as the path to write to stdout.
//
// Output files are created with mode 0600 to protect secret material.
// Parent directories are created automatically with mode 0700.
package export
