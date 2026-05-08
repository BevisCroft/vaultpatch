package normalize

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_UnknownRule(t *testing.T) {
	_, err := New(Rule("invalid"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown rule")
}

func TestNew_ValidRules(t *testing.T) {
	rules := []Rule{RuleLowercase, RuleUppercase, RuleSnakeCase, RuleKebabCase}
	for _, r := range rules {
		_, err := New(r)
		require.NoError(t, err, "rule %q should be valid", r)
	}
}

func TestApply_Lowercase(t *testing.T) {
	n, _ := New(RuleLowercase)
	out, results := n.Apply(map[string]string{
		"DB_HOST": "localhost",
		"api_key": "abc123",
	})
	assert.Equal(t, "localhost", out["db_host"])
	assert.Equal(t, "abc123", out["api_key"])

	changed := findResult(results, "DB_HOST")
	require.NotNil(t, changed)
	assert.True(t, changed.Changed)
	assert.Equal(t, "db_host", changed.Normalized)

	unchanged := findResult(results, "api_key")
	require.NotNil(t, unchanged)
	assert.False(t, unchanged.Changed)
}

func TestApply_Uppercase(t *testing.T) {
	n, _ := New(RuleUppercase)
	out, _ := n.Apply(map[string]string{"db_host": "localhost"})
	assert.Equal(t, "localhost", out["DB_HOST"])
}

func TestApply_SnakeCase(t *testing.T) {
	n, _ := New(RuleSnakeCase)
	cases := map[string]string{
		"camelCase":   "camel_case",
		"kebab-key":   "kebab_key",
		"space key":   "space_key",
		"already_ok":  "already_ok",
	}
	for input, expected := range cases {
		out, _ := n.Apply(map[string]string{input: "v"})
		assert.Equal(t, "v", out[expected], "input: %q", input)
	}
}

func TestApply_KebabCase(t *testing.T) {
	n, _ := New(RuleKebabCase)
	out, _ := n.Apply(map[string]string{"my_key": "val", "camelCase": "val2"})
	assert.Equal(t, "val", out["my-key"])
	assert.Equal(t, "val2", out["camel-case"])
}

func TestApply_DoesNotMutateInput(t *testing.T) {
	n, _ := New(RuleLowercase)
	input := map[string]string{"MY_KEY": "secret"}
	n.Apply(input)
	_, original := input["MY_KEY"]
	assert.True(t, original, "original map should not be mutated")
}

func findResult(results []Result, original string) *Result {
	for i := range results {
		if results[i].Original == original {
			return &results[i]
		}
	}
	return nil
}
