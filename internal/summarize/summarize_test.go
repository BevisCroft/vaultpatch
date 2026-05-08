package summarize_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/example/vaultpatch/internal/diff"
	"github.com/example/vaultpatch/internal/summarize"
)

func changes(path string, ops ...diff.Op) []diff.Change {
	var out []diff.Change
	for i, op := range ops {
		out = append(out, diff.Change{
			Path: path,
			Key:  fmt.Sprintf("key%d", i),
			Op:   op,
		})
	}
	return out
}

func TestBuild_Empty(t *testing.T) {
	s := summarize.New()
	r := s.Build(nil)
	if len(r.Paths) != 0 {
		t.Fatalf("expected no paths, got %d", len(r.Paths))
	}
	if r.TotalAdded+r.TotalRemoved+r.TotalUpdated != 0 {
		t.Fatal("expected zero totals")
	}
}

func TestBuild_SinglePath(t *testing.T) {
	s := summarize.New()
	cs := []diff.Change{
		{Path: "secret/app", Key: "a", Op: diff.OpAdd},
		{Path: "secret/app", Key: "b", Op: diff.OpUpdate},
		{Path: "secret/app", Key: "c", Op: diff.OpRemove},
		{Path: "secret/app", Key: "d", Op: diff.OpNone},
	}
	r := s.Build(cs)
	if len(r.Paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(r.Paths))
	}
	ps := r.Paths[0]
	if ps.Added != 1 || ps.Updated != 1 || ps.Removed != 1 || ps.Unchanged != 1 {
		t.Errorf("unexpected counts: %+v", ps)
	}
	if r.TotalAdded != 1 || r.TotalUpdated != 1 || r.TotalRemoved != 1 {
		t.Errorf("unexpected totals: %+v", r)
	}
}

func TestBuild_MultiplePaths_SortedOrder(t *testing.T) {
	s := summarize.New()
	cs := []diff.Change{
		{Path: "secret/z", Key: "x", Op: diff.OpAdd},
		{Path: "secret/a", Key: "y", Op: diff.OpRemove},
	}
	r := s.Build(cs)
	if len(r.Paths) != 2 {
		t.Fatalf("expected 2 paths")
	}
	if r.Paths[0].Path != "secret/a" {
		t.Errorf("expected sorted first path to be secret/a, got %s", r.Paths[0].Path)
	}
}

func TestFprint_NoChanges(t *testing.T) {
	var buf bytes.Buffer
	summarize.Fprint(&buf, summarize.Report{})
	if !strings.Contains(buf.String(), "No changes") {
		t.Errorf("expected 'No changes' message, got: %s", buf.String())
	}
}

func TestFprint_WithChanges(t *testing.T) {
	var buf bytes.Buffer
	r := summarize.Report{
		Paths: []summarize.PathSummary{
			{Path: "secret/app", Added: 2, Updated: 1, Removed: 0},
		},
		TotalAdded:   2,
		TotalUpdated: 1,
	}
	summarize.Fprint(&buf, r)
	out := buf.String()
	if !strings.Contains(out, "+2 added") {
		t.Errorf("missing added count in output: %s", out)
	}
	if !strings.Contains(out, "~1 updated") {
		t.Errorf("missing updated count in output: %s", out)
	}
	if !strings.Contains(out, "secret/app") {
		t.Errorf("missing path in output: %s", out)
	}
}
