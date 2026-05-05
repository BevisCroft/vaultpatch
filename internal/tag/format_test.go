package tag_test

import (
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/tag"
)

func TestFprint_DryRun(t *testing.T) {
	results := []tag.Result{
		{Path: "app/config", Tags: map[string]string{"env": "staging"}, DryRun: true},
	}
	var buf strings.Builder
	tag.Fprint(&buf, results)
	if !strings.Contains(buf.String(), "dry-run") {
		t.Errorf("expected dry-run label, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "env=staging") {
		t.Errorf("expected tag key=value, got: %s", buf.String())
	}
}

func TestFprint_Error(t *testing.T) {
	results := []tag.Result{
		{Path: "app/config", Err: errSentinel("boom")},
	}
	var buf strings.Builder
	tag.Fprint(&buf, results)
	if !strings.Contains(buf.String(), "ERROR") {
		t.Errorf("expected ERROR label, got: %s", buf.String())
	}
}

func TestFprint_MultipleTags(t *testing.T) {
	results := []tag.Result{
		{Path: "svc/db", Tags: map[string]string{"env": "prod", "owner": "dba"}, DryRun: false},
	}
	var buf strings.Builder
	tag.Fprint(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "env=prod") || !strings.Contains(out, "owner=dba") {
		t.Errorf("expected both tags in output, got: %s", out)
	}
}

func TestFprintList_NoTags(t *testing.T) {
	var buf strings.Builder
	tag.FprintList(&buf, "app/cfg", map[string]string{})
	if !strings.Contains(buf.String(), "no tags") {
		t.Errorf("expected no tags message, got: %s", buf.String())
	}
}

func TestFprintList_WithTags(t *testing.T) {
	var buf strings.Builder
	tag.FprintList(&buf, "app/cfg", map[string]string{"team": "platform"})
	if !strings.Contains(buf.String(), "team=platform") {
		t.Errorf("expected tag in output, got: %s", buf.String())
	}
}

type errSentinel string

func (e errSentinel) Error() string { return string(e) }
