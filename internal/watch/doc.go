// Package watch implements a polling-based watcher for HashiCorp Vault secret
// paths. It periodically reads configured paths and emits change events when
// the secret data differs from the previously observed state.
//
// # Usage
//
//	w := watch.New(client, "secret", []string{"myapp/config"}, 30*time.Second)
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	for event := range w.Watch(ctx) {
//		fmt.Printf("change at %s: %d key(s) changed\n", event.Path, len(event.Changes))
//	}
//
// The watcher runs until the provided context is cancelled. The first read of
// each path establishes the baseline and does not produce an event.
package watch
