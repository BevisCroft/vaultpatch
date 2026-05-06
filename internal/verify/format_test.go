package verify_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/verify"
)

func TestFprint_AllPass(t *testing.T) {
	results := []verify.Result{
		{Path: "app/cfg", Key: "host", Expected: "localhost", Actual: "localhost", Match: true},
	}
	var buf bytes.Buffer
	verify.Fprint(&buf, results, false)
	if !strings.Contains(buf.String(), "[PASS]") {
		t.Errorf("expected PASS in output, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "1 passed") {
		t.Errorf("expected summary, got: %s", buf.String())
	}
}

func TestFprint_Failure(t *testing.T) {
	results := []verify.Result{
		{Path: "app/cfg", Key: "host", Expected: "localhost", Actual: "remote", Match: false},
	}
	var buf bytes.Buffer
	verify.Fprint(&buf, results, false)
	if !strings.Contains(buf.String(), "[FAIL]") {
		t.Errorf("expected FAIL in output, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "localhost") {
		t.Errorf("expected expected value in output")
	}
}

func TestFprint_MaskValues(t *testing.T) {
	results := []verify.Result{
		{Path: "app/cfg", Key: "secret", Expected: "s3cr3t", Actual: "wrong", Match: false},
	}
	var buf bytes.Buffer
	verify.Fprint(&buf, results, true)
	if strings.Contains(buf.String(), "s3cr3t") {
		t.Error("expected value should be masked")
	}
	if !strings.Contains(buf.String(), "***") {
		t.Error("expected mask placeholder '***'")
	}
}

func TestFprint_Error(t *testing.T) {
	results := []verify.Result{
		{Path: "app/cfg", Err: errors.New("connection refused")},
	}
	var buf bytes.Buffer
	verify.Fprint(&buf, results, false)
	if !strings.Contains(buf.String(), "[ERROR]") {
		t.Errorf("expected ERROR in output, got: %s", buf.String())
	}
}

func TestFprint_Empty(t *testing.T) {
	var buf bytes.Buffer
	verify.Fprint(&buf, nil, false)
	if !strings.Contains(buf.String(), "no expectations") {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}
