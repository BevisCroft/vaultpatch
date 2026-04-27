package snapshot_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/your-org/vaultpatch/internal/snapshot"
)

// stubLister is a test double implementing snapshot.Lister.
type stubLister struct {
	paths   []string
	secrets map[string]map[string]string
}

func (s *stubLister) ListSecrets(_ context.Context, _ string) ([]string, error) {
	return s.paths, nil
}

func (s *stubLister) ReadSecret(_ context.Context, path string) (map[string]string, error) {
	return s.secrets[path], nil
}

func TestCapture_PopulatesSecrets(t *testing.T) {
	client := &stubLister{
		paths: []string{"app/db", "app/api"},
		secrets: map[string]map[string]string{
			"app/db":  {"password": "s3cr3t"},
			"app/api": {"key": "abc123"},
		},
	}

	snap, err := snapshot.Capture(context.Background(), client, "secret", "app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if snap.Mount != "secret" {
		t.Errorf("mount: got %q, want %q", snap.Mount, "secret")
	}
	if len(snap.Secrets) != 2 {
		t.Errorf("secrets count: got %d, want 2", len(snap.Secrets))
	}
	if snap.Secrets["app/db"]["password"] != "s3cr3t" {
		t.Errorf("expected password s3cr3t, got %q", snap.Secrets["app/db"]["password"])
	}
	if snap.CapturedAt.IsZero() {
		t.Error("CapturedAt should not be zero")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	original := &snapshot.Snapshot{
		CapturedAt: time.Now().UTC().Truncate(time.Second),
		Mount:      "kv",
		BasePath:   "svc",
		Secrets: map[string]snapshot.SecretMap{
			"svc/auth": {"token": "xyz"},
		},
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "snap.json")

	if err := snapshot.Save(original, filePath); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := snapshot.Load(filePath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Mount != original.Mount {
		t.Errorf("mount mismatch: got %q, want %q", loaded.Mount, original.Mount)
	}
	if loaded.Secrets["svc/auth"]["token"] != "xyz" {
		t.Errorf("token mismatch after round-trip")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := snapshot.Load("/nonexistent/path/snap.json")
	if err == nil {
		t.Error("expected error loading missing file, got nil")
	}
}

func TestSave_InvalidPath(t *testing.T) {
	snap := &snapshot.Snapshot{Secrets: make(map[string]snapshot.SecretMap)}
	err := snapshot.Save(snap, "/no/such/dir/snap.json")
	if err == nil {
		t.Error("expected error saving to invalid path, got nil")
	}
	_ = os.Remove("/no/such/dir/snap.json") // no-op, just safe cleanup
}
