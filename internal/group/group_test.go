package group_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/example/vaultpatch/internal/group"
)

func TestApply_SingleDepth(t *testing.T) {
	g := group.New(group.Options{Depth: 1})
	paths := []string{
		"secret/app/db",
		"secret/app/api",
		"secret/infra/redis",
	}

	results := g.Apply(paths)

	if len(results) != 1 {
		t.Fatalf("expected 1 group, got %d", len(results))
	}
	if results[0].Prefix != "secret" {
		t.Errorf("expected prefix 'secret', got %q", results[0].Prefix)
	}
	if len(results[0].Paths) != 3 {
		t.Errorf("expected 3 paths in group, got %d", len(results[0].Paths))
	}
}

func TestApply_TwoDepth(t *testing.T) {
	g := group.New(group.Options{Depth: 2})
	paths := []string{
		"secret/app/db",
		"secret/app/api",
		"secret/infra/redis",
	}

	results := g.Apply(paths)

	if len(results) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(results))
	}
	if results[0].Prefix != "secret/app" {
		t.Errorf("expected first prefix 'secret/app', got %q", results[0].Prefix)
	}
	if results[1].Prefix != "secret/infra" {
		t.Errorf("expected second prefix 'secret/infra', got %q", results[1].Prefix)
	}
}

func TestApply_DefaultDepthIsOne(t *testing.T) {
	g := group.New(group.Options{})
	paths := []string{"a/b/c", "a/d/e"}
	results := g.Apply(paths)
	if len(results) != 1 || results[0].Prefix != "a" {
		t.Errorf("expected single group 'a', got %+v", results)
	}
}

func TestApply_EmptyInput(t *testing.T) {
	g := group.New(group.Options{Depth: 1})
	results := g.Apply(nil)
	if len(results) != 0 {
		t.Errorf("expected no results, got %d", len(results))
	}
}

func TestApply_SortedOutput(t *testing.T) {
	g := group.New(group.Options{Depth: 1})
	paths := []string{"z/key", "a/key", "m/key"}
	results := g.Apply(paths)
	if results[0].Prefix != "a" || results[1].Prefix != "m" || results[2].Prefix != "z" {
		t.Errorf("results not sorted: %+v", results)
	}
}

func TestFprint_NoResults(t *testing.T) {
	var buf bytes.Buffer
	group.Fprint(&buf, nil)
	if !strings.Contains(buf.String(), "no paths") {
		t.Errorf("expected 'no paths' message, got %q", buf.String())
	}
}

func TestFprint_WithResults(t *testing.T) {
	g := group.New(group.Options{Depth: 2})
	paths := []string{"secret/app/db", "secret/app/api"}
	results := g.Apply(paths)

	var buf bytes.Buffer
	group.Fprint(&buf, results)
	out := buf.String()

	if !strings.Contains(out, "secret/app") {
		t.Errorf("expected group prefix in output, got %q", out)
	}
	if !strings.Contains(out, "2 path") {
		t.Errorf("expected path count in output, got %q", out)
	}
}
