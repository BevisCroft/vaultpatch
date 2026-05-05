package tag_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/tag"
)

type stubWriter struct {
	metadata  map[string]map[string]string
	writeErr  error
	readErr   error
}

func (s *stubWriter) ReadSecretMetadata(_ context.Context, path string) (map[string]string, error) {
	if s.readErr != nil {
		return nil, s.readErr
	}
	if m, ok := s.metadata[path]; ok {
		return m, nil
	}
	return map[string]string{}, nil
}

func (s *stubWriter) WriteSecretMetadata(_ context.Context, path string, metadata map[string]string) error {
	if s.writeErr != nil {
		return s.writeErr
	}
	if s.metadata == nil {
		s.metadata = make(map[string]map[string]string)
	}
	s.metadata[path] = metadata
	return nil
}

func TestApply_DryRun_NoWrites(t *testing.T) {
	stub := &stubWriter{}
	m := tag.New(stub, "secret", true)
	results := m.Apply(context.Background(), []string{"app/cfg"}, map[string]string{"env": "staging"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if len(stub.metadata) != 0 {
		t.Fatal("dry-run should not write metadata")
	}
}

func TestApply_MergesExistingTags(t *testing.T) {
	stub := &stubWriter{
		metadata: map[string]map[string]string{
			"secret/app/cfg": {"owner": "team-a"},
		},
	}
	m := tag.New(stub, "secret", false)
	m.Apply(context.Background(), []string{"app/cfg"}, map[string]string{"env": "prod"})
	got := stub.metadata["secret/app/cfg"]
	if got["owner"] != "team-a" {
		t.Errorf("expected existing tag to be preserved, got %v", got)
	}
	if got["env"] != "prod" {
		t.Errorf("expected new tag to be set, got %v", got)
	}
}

func TestApply_WriteError_CapturedInResult(t *testing.T) {
	stub := &stubWriter{writeErr: errors.New("vault unavailable")}
	m := tag.New(stub, "secret", false)
	results := m.Apply(context.Background(), []string{"app/cfg"}, map[string]string{"env": "prod"})
	if results[0].Err == nil {
		t.Fatal("expected error to be captured in result")
	}
}

func TestList_ReturnsTags(t *testing.T) {
	stub := &stubWriter{
		metadata: map[string]map[string]string{
			"secret/app/cfg": {"env": "prod", "owner": "ops"},
		},
	}
	m := tag.New(stub, "secret", false)
	tags, err := m.List(context.Background(), "app/cfg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tags["env"] != "prod" {
		t.Errorf("expected env=prod, got %v", tags["env"])
	}
}

func TestSortedKeys_DeterministicOrder(t *testing.T) {
	tags := map[string]string{"z": "1", "a": "2", "m": "3"}
	keys := tag.SortedKeys(tags)
	expected := []string{"a", "m", "z"}
	for i, k := range keys {
		if k != expected[i] {
			t.Errorf("index %d: want %s got %s", i, expected[i], k)
		}
	}
}
