package flatten_test

import (
	"testing"

	"github.com/your-org/vaultpatch/internal/flatten"
)

func TestApply_FlatInput(t *testing.T) {
	f := flatten.New(".")
	input := map[string]any{
		"host": "localhost",
		"port": "5432",
	}
	res := f.Apply(input)
	if len(res.Warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", res.Warnings)
	}
	assertEqual(t, "localhost", res.Flat["host"])
	assertEqual(t, "5432", res.Flat["port"])
}

func TestApply_NestedMap(t *testing.T) {
	f := flatten.New(".")
	input := map[string]any{
		"db": map[string]any{
			"host": "pg.internal",
			"port": "5432",
		},
	}
	res := f.Apply(input)
	if len(res.Warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", res.Warnings)
	}
	assertEqual(t, "pg.internal", res.Flat["db.host"])
	assertEqual(t, "5432", res.Flat["db.port"])
}

func TestApply_DeeplyNested(t *testing.T) {
	f := flatten.New("/")
	input := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "deep",
			},
		},
	}
	res := f.Apply(input)
	assertEqual(t, "deep", res.Flat["a/b/c"])
}

func TestApply_SliceProducesWarning(t *testing.T) {
	f := flatten.New(".")
	input := map[string]any{
		"items": []any{"a", "b"},
	}
	res := f.Apply(input)
	if len(res.Warnings) == 0 {
		t.Fatal("expected warning for slice, got none")
	}
	if _, ok := res.Flat["items"]; ok {
		t.Fatal("slice key should not appear in flat output")
	}
}

func TestApply_NilValue(t *testing.T) {
	f := flatten.New(".")
	input := map[string]any{"key": nil}
	res := f.Apply(input)
	assertEqual(t, "", res.Flat["key"])
}

func TestApply_NonStringScalar(t *testing.T) {
	f := flatten.New(".")
	input := map[string]any{"count": 42}
	res := f.Apply(input)
	assertEqual(t, "42", res.Flat["count"])
}

func TestExpand_RoundTrip(t *testing.T) {
	f := flatten.New(".")
	original := map[string]any{
		"db": map[string]any{
			"host": "localhost",
			"port": "5432",
		},
	}
	flat := f.Apply(original)
	if len(flat.Warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", flat.Warnings)
	}
	expanded := f.Expand(flat.Flat)
	if len(expanded.Warnings) != 0 {
		t.Fatalf("unexpected expand warnings: %v", expanded.Warnings)
	}
}

func TestExpand_ConflictWarning(t *testing.T) {
	f := flatten.New(".")
	// "a" is both a leaf and a namespace prefix
	input := map[string]string{
		"a":   "leaf",
		"a.b": "nested",
	}
	res := f.Expand(input)
	if len(res.Warnings) == 0 {
		t.Fatal("expected conflict warning, got none")
	}
}

func TestNew_DefaultSeparator(t *testing.T) {
	f := flatten.New("")
	input := map[string]any{"x": map[string]any{"y": "z"}}
	res := f.Apply(input)
	if _, ok := res.Flat["x.y"]; !ok {
		t.Fatal("expected key 'x.y' with default separator")
	}
}

func assertEqual(t *testing.T, want, got string) {
	t.Helper()
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}
