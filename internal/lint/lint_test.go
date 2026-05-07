package lint_test

import (
	"testing"

	"github.com/yourusername/vaultpatch/internal/lint"
)

func defaultLinter() *lint.Linter {
	return lint.New(lint.DefaultRules())
}

func TestRun_NoFindings(t *testing.T) {
	l := defaultLinter()
	findings := l.Run("secret/app/prod", map[string]string{
		"database_host": "db.example.com",
		"port":          "5432",
	})
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %d: %+v", len(findings), findings)
	}
}

func TestRun_UppercaseKeyWarning(t *testing.T) {
	l := defaultLinter()
	findings := l.Run("secret/app", map[string]string{
		"DB_HOST": "localhost",
	})
	if !containsSeverity(findings, lint.SeverityWarning) {
		t.Fatal("expected a warning for uppercase key")
	}
}

func TestRun_EmptyValueError(t *testing.T) {
	l := defaultLinter()
	findings := l.Run("secret/app", map[string]string{
		"api_url": "",
	})
	if !containsSeverity(findings, lint.SeverityError) {
		t.Fatal("expected an error for empty value")
	}
}

func TestRun_WhitespaceOnlyValueError(t *testing.T) {
	l := defaultLinter()
	findings := l.Run("secret/app", map[string]string{
		"region": "   ",
	})
	if !containsSeverity(findings, lint.SeverityError) {
		t.Fatal("expected an error for whitespace-only value")
	}
}

func TestRun_WhitespaceInKeyError(t *testing.T) {
	l := defaultLinter()
	findings := l.Run("secret/app", map[string]string{
		"my key": "value",
	})
	if !containsSeverity(findings, lint.SeverityError) {
		t.Fatal("expected an error for key with whitespace")
	}
}

func TestRun_SensitiveKeyNameWarning(t *testing.T) {
	l := defaultLinter()
	findings := l.Run("secret/app", map[string]string{
		"db_password": "hunter2",
	})
	if !containsSeverity(findings, lint.SeverityWarning) {
		t.Fatal("expected a warning for sensitive-looking key name")
	}
}

func TestRun_CustomRule(t *testing.T) {
	customRule := lint.Rule{
		Name: "no-localhost",
		Check: func(path, key, value string) *lint.Finding {
			if value == "localhost" {
				return &lint.Finding{
					Path:     path,
					Key:      key,
					Message:  "value is 'localhost'; not suitable for production",
					Severity: lint.SeverityWarning,
				}
			}
			return nil
		},
	}
	l := lint.New([]lint.Rule{customRule})
	findings := l.Run("secret/app", map[string]string{
		"db_host": "localhost",
	})
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Severity != lint.SeverityWarning {
		t.Errorf("expected warning, got %s", findings[0].Severity)
	}
}

func containsSeverity(findings []lint.Finding, s lint.Severity) bool {
	for _, f := range findings {
		if f.Severity == s {
			return true
		}
	}
	return false
}
