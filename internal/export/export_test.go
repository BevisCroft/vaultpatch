package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNew_InvalidFormat(t *testing.T) {
	_, err := New("toml")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestNew_ValidFormats(t *testing.T) {
	for _, f := range []string{"json", "yaml", "env", "JSON", "YAML"} {
		_, err := New(f)
		if err != nil {
			t.Fatalf("unexpected error for format %q: %v", f, err)
		}
	}
}

func TestWrite_JSON(t *testing.T) {
	e, _ := New("json")
	secrets := map[string]string{"key": "value", "db_pass": "s3cr3t"}
	path := filepath.Join(t.TempDir(), "out.json")
	if err := e.Write(secrets, path); err != nil {
		t.Fatalf("Write: %v", err)
	}
	data, _ := os.ReadFile(path)
	var got map[string]string
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["key"] != "value" {
		t.Errorf("got key=%q, want \"value\"", got["key"])
	}
}

func TestWrite_YAML(t *testing.T) {
	e, _ := New("yaml")
	secrets := map[string]string{"foo": "bar"}
	path := filepath.Join(t.TempDir(), "out.yaml")
	if err := e.Write(secrets, path); err != nil {
		t.Fatalf("Write: %v", err)
	}
	data, _ := os.ReadFile(path)
	var got map[string]string
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["foo"] != "bar" {
		t.Errorf("got foo=%q, want \"bar\"", got["foo"])
	}
}

func TestWrite_Env(t *testing.T) {
	e, _ := New("env")
	secrets := map[string]string{"api_key": "abc123"}
	path := filepath.Join(t.TempDir(), "out.env")
	if err := e.Write(secrets, path); err != nil {
		t.Fatalf("Write: %v", err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "API_KEY=") {
		t.Errorf("env output missing API_KEY, got: %s", data)
	}
}

func TestWrite_CreatesParentDir(t *testing.T) {
	e, _ := New("json")
	path := filepath.Join(t.TempDir(), "sub", "dir", "out.json")
	if err := e.Write(map[string]string{"x": "y"}, path); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("output file not created: %v", err)
	}
}
