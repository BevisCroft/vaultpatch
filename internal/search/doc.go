// Package search provides full-text search across Vault secret paths.
//
// It walks all secrets under a configured KV mount and returns results whose
// key names (and optionally values) match the supplied query string.
//
// Usage:
//
//	s := search.New(client, "secret")
//	results, err := s.Find(ctx, "api_key", false)
//	if err != nil {
//		log.Fatal(err)
//	}
//	search.Fprint(os.Stdout, results, true)
package search
