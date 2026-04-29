// Package history records a persistent, append-only log of every apply
// operation performed by vaultpatch.
//
// Each invocation of [Recorder.Record] appends a newline-delimited JSON entry
// to the configured file, capturing metadata such as the target environment,
// operator identity, dry-run flag, change count, and failure count.
//
// Entries can be replayed at any time with [Load], which returns all recorded
// entries in chronological order.
//
// Typical usage:
//
//	r := history.New("/var/log/vaultpatch/history.jsonl")
//	err := r.Record(history.Entry{
//		Environment: "production",
//		Operator:    os.Getenv("USER"),
//		Changes:     applied,
//		Failures:    failed,
//	})
package history
