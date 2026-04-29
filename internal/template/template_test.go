package template

import (
	"strings"
	"testing"
)

func testSecrets() map[string]map[string]string {
	return map[string]map[string]string{
		"secret/db": {
			"username": "admin",
			"password": "s3cr3t",
		},
		"secret/app": {
			"api_key": "abc123",
		},
	}
}

func TestRender_BasicTemplate(t *testing.T) {
	r := New(testSecrets())
	out, err := r.Render(`user={{ secret "secret/db" "username" }}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "user=admin" {
		t.Errorf("got %q, want %q", out, "user=admin")
	}
}

func TestRender_MultipleSecrets(t *testing.T) {
	r := New(testSecrets())
	tmpl := `DSN={{ secret "secret/db" "username" }}:{{ secret "secret/db" "password" }}@localhost`
	out, err := r.Render(tmpl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "DSN=admin:s3cr3t@localhost"
	if out != expected {
		t.Errorf("got %q, want %q", out, expected)
	}
}

func TestRender_DotSecretsAccess(t *testing.T) {
	r := New(testSecrets())
	out, err := r.Render(`key={{ index (index .Secrets "secret/app") "api_key" }}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "key=abc123" {
		t.Errorf("got %q, want %q", out, "key=abc123")
	}
}

func TestRender_UnknownPath(t *testing.T) {
	r := New(testSecrets())
	_, err := r.Render(`{{ secret "secret/missing" "key" }}`)
	if err == nil {
		t.Fatal("expected error for missing path, got nil")
	}
	if !strings.Contains(err.Error(), "secret/missing") {
		t.Errorf("error should mention missing path, got: %v", err)
	}
}

func TestRender_UnknownKey(t *testing.T) {
	r := New(testSecrets())
	_, err := r.Render(`{{ secret "secret/db" "nokey" }}`)
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
	if !strings.Contains(err.Error(), "nokey") {
		t.Errorf("error should mention missing key, got: %v", err)
	}
}

func TestRender_InvalidTemplateSyntax(t *testing.T) {
	r := New(testSecrets())
	_, err := r.Render(`{{ unclosed`)
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
}
