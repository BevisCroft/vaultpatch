package digest_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exampleorg/vaultpatch/internal/digest"
	"github.com/exampleorg/vaultpatch/internal/vault"
)

// stubReader satisfies digest.SecretReader in-process.
type stubReader struct {
	data map[string]map[string]string
	err  error
}

func (s *stubReader) ReadSecret(path string) (map[string]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	if m, ok := s.data[path]; ok {
		return m, nil
	}
	return nil, errors.New("not found")
}

func newMockVault(t *testing.T, secrets map[string]map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
}

func newTestClient(t *testing.T, srv *httptest.Server) *vault.Client {
	t.Helper()
	c, err := vault.NewClient(srv.URL, "test-token", "secret")
	require.NoError(t, err)
	return c
}

func TestCompute_StableDigest(t *testing.T) {
	stub := &stubReader{
		data: map[string]map[string]string{
			"secret/data/app": {"key": "value", "foo": "bar"},
		},
	}
	d := digest.New(stub)

	r1 := d.Compute("secret/data/app")
	r2 := d.Compute("secret/data/app")

	require.NoError(t, r1.Err)
	assert.Equal(t, r1.Digest, r2.Digest, "digest must be deterministic")
	assert.Len(t, r1.Digest, 64, "SHA-256 hex string should be 64 chars")
}

func TestCompute_OrderIndependent(t *testing.T) {
	stub1 := &stubReader{data: map[string]map[string]string{
		"p": {"a": "1", "b": "2"},
	}}
	stub2 := &stubReader{data: map[string]map[string]string{
		"p": {"b": "2", "a": "1"},
	}}

	d1 := digest.New(stub1).Compute("p")
	d2 := digest.New(stub2).Compute("p")
	assert.Equal(t, d1.Digest, d2.Digest)
}

func TestCompute_ErrorPropagates(t *testing.T) {
	stub := &stubReader{err: errors.New("vault unavailable")}
	d := digest.New(stub)
	r := d.Compute("any/path")
	require.Error(t, r.Err)
	assert.Empty(t, r.Digest)
}

func TestCompare_MatchTrue(t *testing.T) {
	stub := &stubReader{data: map[string]map[string]string{
		"p": {"x": "y"},
	}}
	d := digest.New(stub)
	baseline := d.Compute("p").Digest

	r := d.Compare("p", baseline)
	assert.True(t, r.Match)
}

func TestCompare_MatchFalse(t *testing.T) {
	stub := &stubReader{data: map[string]map[string]string{
		"p": {"x": "changed"},
	}}
	d := digest.New(stub)
	r := d.Compare("p", "deadbeef")
	assert.False(t, r.Match)
}

func TestFprint_Results(t *testing.T) {
	results := []digest.Result{
		{Path: "secret/a", Digest: strings.Repeat("a", 64), Match: true},
		{Path: "secret/b", Digest: strings.Repeat("b", 64), Match: false},
		{Path: "secret/c", Err: errors.New("boom")},
	}
	var buf bytes.Buffer
	digest.Fprint(&buf, results, false)
	out := buf.String()
	assert.Contains(t, out, "secret/a")
	assert.Contains(t, out, "secret/b")
	assert.Contains(t, out, "boom")
}

func TestFprint_Empty(t *testing.T) {
	var buf bytes.Buffer
	digest.Fprint(&buf, nil, false)
	assert.Contains(t, buf.String(), "no paths")
}
