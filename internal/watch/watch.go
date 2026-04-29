// Package watch provides functionality for monitoring Vault secret paths
// for changes over a configurable polling interval.
package watch

import (
	"context"
	"time"

	"github.com/your-org/vaultpatch/internal/diff"
	"github.com/your-org/vaultpatch/internal/vault"
)

// Event represents a detected change at a secret path.
type Event struct {
	Path    string
	Changes []diff.Change
	At      time.Time
}

// Watcher polls Vault secret paths and emits events when changes are detected.
type Watcher struct {
	client   *vault.Client
	paths    []string
	interval time.Duration
	mount    string
}

// New creates a new Watcher for the given paths and poll interval.
func New(client *vault.Client, mount string, paths []string, interval time.Duration) *Watcher {
	return &Watcher{
		client:   client,
		paths:    paths,
		interval: interval,
		mount:    mount,
	}
}

// Watch starts polling and sends Events to the returned channel.
// It stops when ctx is cancelled. The channel is closed on exit.
func (w *Watcher) Watch(ctx context.Context) <-chan Event {
	ch := make(chan Event, len(w.paths))

	go func() {
		defer close(ch)

		prev := make(map[string]map[string]string)

		for {
			for _, p := range w.paths {
				current, err := w.client.ReadSecret(ctx, w.mount, p)
				if err != nil {
					continue
				}
				old, seen := prev[p]
				if !seen {
					prev[p] = current
					continue
				}
				changes := diff.Compute(old, current)
				if len(changes) > 0 {
					select {
					case ch <- Event{Path: p, Changes: changes, At: time.Now()}:
					case <-ctx.Done():
						return
					}
					prev[p] = current
				}
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(w.interval):
			}
		}
	}()

	return ch
}
