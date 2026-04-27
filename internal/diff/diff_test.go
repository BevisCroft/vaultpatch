package diff_test

import (
	"testing"

	"github.com/your-org/vaultpatch/internal/diff"
)

func TestCompute_NoChanges(t *testing.T) {
	src := map[string]string{"key1": "val1", "key2": "val2"}
	dst := map[string]string{"key1": "val1", "key2": "val2"}

	result := diff.Compute(src, dst)
	if result.HasChanges() {
		t.Errorf("expected no changes, got %d", len(result.Changes))
	}
}

func TestCompute_Added(t *testing.T) {
	src := map[string]string{}
	dst := map[string]string{"newkey": "newval"}

	result := diff.Compute(src, dst)
	if len(result.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result.Changes))
	}
	c := result.Changes[0]
	if c.Type != diff.ChangeAdded || c.Key != "newkey" || c.NewValue != "newval" {
		t.Errorf("unexpected change: %+v", c)
	}
}

func TestCompute_Removed(t *testing.T) {
	src := map[string]string{"oldkey": "oldval"}
	dst := map[string]string{}

	result := diff.Compute(src, dst)
	if len(result.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result.Changes))
	}
	c := result.Changes[0]
	if c.Type != diff.ChangeRemoved || c.Key != "oldkey" || c.OldValue != "oldval" {
		t.Errorf("unexpected change: %+v", c)
	}
}

func TestCompute_Updated(t *testing.T) {
	src := map[string]string{"key": "old"}
	dst := map[string]string{"key": "new"}

	result := diff.Compute(src, dst)
	if len(result.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result.Changes))
	}
	c := result.Changes[0]
	if c.Type != diff.ChangeUpdated || c.OldValue != "old" || c.NewValue != "new" {
		t.Errorf("unexpected change: %+v", c)
	}
}

func TestCompute_SortedOutput(t *testing.T) {
	src := map[string]string{"z": "1", "a": "1"}
	dst := map[string]string{"z": "2", "a": "2"}

	result := diff.Compute(src, dst)
	if len(result.Changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(result.Changes))
	}
	if result.Changes[0].Key != "a" || result.Changes[1].Key != "z" {
		t.Errorf("expected sorted keys, got %s, %s", result.Changes[0].Key, result.Changes[1].Key)
	}
}
