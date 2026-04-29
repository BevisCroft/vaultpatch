// Package drift detects configuration drift by comparing a previously
// captured snapshot of Vault secrets against the current live state.
//
// Typical usage:
//
//	snap, _ := snapshot.Load("snapshot.json")
//	detector := drift.New(client, "secret")
//	reports, err := detector.Detect(ctx, snap)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	drift.Fprint(os.Stdout, reports, false)
//
// A Report is produced for every Vault path whose live key/value pairs
// differ from the snapshot. Paths present in the snapshot but missing
// from Vault are reported as full removals.
package drift
