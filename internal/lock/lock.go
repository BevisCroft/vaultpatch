// Package lock provides a simple advisory locking mechanism for Vault paths
// to prevent concurrent modifications during patch or promote operations.
package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
)

const (
	defaultLockTTL  = 30 * time.Second
	lockMetaKey     = "_vaultpatch_lock"
	lockMetaOwner   = "_vaultpatch_owner"
)

// Locker acquires and releases advisory locks stored as metadata in Vault KV.
type Locker struct {
	client *api.Client
	mount  string
	owner  string
}

// New creates a new Locker for the given Vault client, KV mount, and owner
// identifier (e.g. hostname or CI job ID).
func New(client *api.Client, mount, owner string) *Locker {
	return &Locker{client: client, mount: mount, owner: owner}
}

// Acquire writes a lock entry at the given path. It returns an error if the
// path is already locked by a different owner.
func (l *Locker) Acquire(ctx context.Context, path string) error {
	existing, err := l.readLock(ctx, path)
	if err != nil {
		return fmt.Errorf("lock: read existing: %w", err)
	}
	if existing != "" && existing != l.owner {
		return fmt.Errorf("lock: path %q is locked by %q", path, existing)
	}

	data := map[string]interface{}{
		lockMetaKey:   time.Now().UTC().Format(time.RFC3339),
		lockMetaOwner: l.owner,
	}
	kv := l.client.KVv2(l.mount)
	if _, err := kv.Put(ctx, path+"/.lock", data); err != nil {
		return fmt.Errorf("lock: acquire %q: %w", path, err)
	}
	return nil
}

// Release removes the lock entry for the given path. It is a no-op if the
// lock is not held by this owner.
func (l *Locker) Release(ctx context.Context, path string) error {
	existing, err := l.readLock(ctx, path)
	if err != nil {
		return fmt.Errorf("lock: read for release: %w", err)
	}
	if existing == "" || existing != l.owner {
		return nil
	}
	kv := l.client.KVv2(l.mount)
	if err := kv.Delete(ctx, path+"/.lock"); err != nil {
		return fmt.Errorf("lock: release %q: %w", path, err)
	}
	return nil
}

func (l *Locker) readLock(ctx context.Context, path string) (string, error) {
	kv := l.client.KVv2(l.mount)
	secret, err := kv.Get(ctx, path+"/.lock")
	if err != nil {
		// Treat a missing secret as no lock held.
		return "", nil
	}
	if secret == nil || secret.Data == nil {
		return "", nil
	}
	owner, _ := secret.Data[lockMetaOwner].(string)
	return owner, nil
}
