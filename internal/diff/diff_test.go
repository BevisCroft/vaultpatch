package diff

import (
	"bytes"
	"strings"
	"testing"
)

func TestCompute_NoChanges(t *testing.T) {
	secrets := map[string]string{"key": "value"}
	entries := Compute(secrets, secrets)
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestCompute_Added(t *testing.T) {
	old := map[string]string{}
	new := map[string]string{"foo": "bar"}
	entries := Compute(old, new)
	if len(entries) != 1 || entries[0].Op != OpAdded {
		t.Fatalf("expected 1 added entry, got %+v", entries)
	}
	if entries[0].Key != "foo" || entries[0].NewValue != "bar" {
		t.Errorf("unexpected entry: %+v", entries[0])
	}
}

func TestCompute_Removed(t *testing.T) {
	old := map[string]string{"foo": "bar"}
	new := map[string]string{}
	entries := Compute(old, new)
	if len(entries) != 1 || entries[0].Op != OpRemoved {
		t.Fatalf("expected 1 removed entry, got %+v", entries)
	}
	if entries[0].OldValue != "bar" {
		t.Errorf("unexpected old value: %s", entries[0].OldValue)
	}
}

func TestCompute_Updated(t *testing.T) {
	old := map[string]string{"foo": "old"}
	new := map[string]string{"foo": "new"}
	entries := Compute(old, new)
	if len(entries) != 1 || entries[0].Op != OpUpdated {
		t.Fatalf("expected 1 updated entry, got %+v", entries)
	}
	if entries[0].OldValue != "old" || entries[0].NewValue != "new" {
		t.Errorf("unexpected values: %+v", entries[0])
	}
}

func TestCompute_SortedOutput(t *testing.T) {
	old := map[string]string{}
	new := map[string]string{"z": "1", "a": "2", "m": "3"}
	entries := Compute(old, new)
	keys := []string{entries[0].Key, entries[1].Key, entries[2].Key}
	if keys[0] != "a" || keys[1] != "m" || keys[2] != "z" {
		t.Errorf("expected sorted keys, got %v", keys)
	}
}

func TestFprint_NoChanges(t *testing.T) {
	var buf bytes.Buffer
	Fprint(&buf, []Entry{}, false)
	if !strings.Contains(buf.String(), "No changes") {
		t.Errorf("expected no-changes message, got: %s", buf.String())
	}
}

func TestFprint_MaskSecrets(t *testing.T) {
	var buf bytes.Buffer
	entries := []Entry{{Key: "password", NewValue: "s3cr3t", Op: OpAdded}}
	Fprint(&buf, entries, true)
	if strings.Contains(buf.String(), "s3cr3t") {
		t.Errorf("expected secret to be masked, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "***") {
		t.Errorf("expected mask placeholder, got: %s", buf.String())
	}
}

func TestFprint_Summary(t *testing.T) {
	var buf bytes.Buffer
	entries := []Entry{
		{Key: "a", NewValue: "1", Op: OpAdded},
		{Key: "b", OldValue: "2", Op: OpRemoved},
		{Key: "c", OldValue: "old", NewValue: "new", Op: OpUpdated},
	}
	Fprint(&buf, entries, false)
	out := buf.String()
	if !strings.Contains(out, "1 added") || !strings.Contains(out, "1 removed") || !strings.Contains(out, "1 updated") {
		t.Errorf("unexpected summary: %s", out)
	}
}
