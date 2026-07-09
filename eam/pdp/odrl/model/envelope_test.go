package model

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseEnvelope_RawTurtle covers B3: a bare Turtle policy document (as
// import.sh copies a bronhouder beleid.ttl into the store) must load without a
// hand-crafted JSON envelope. ParseEnvelope wraps it on the fly.
func TestParseEnvelope_RawTurtle(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join(examples, "1-generiek-drietraps.ttl"))
	require.NoError(t, err)

	env, err := ParseEnvelope(raw)
	require.NoError(t, err)
	require.NotNil(t, env)
	assert.Equal(t, mimeTurtle, env.MimeType)
	require.NotNil(t, env.Document)
	require.Len(t, env.Document.Policies, 3)

	// the raw source round-trips byte-for-byte.
	src, err := env.RawSource()
	require.NoError(t, err)
	assert.Equal(t, raw, src)
}

// TestParseEnvelope_JSONEnvelope keeps the original storage format working.
func TestParseEnvelope_JSONEnvelope(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join(examples, "1-generiek-drietraps.ttl"))
	require.NoError(t, err)

	doc, err := Parse(bytes.NewReader(raw), mimeTurtle)
	require.NoError(t, err)

	data, err := NewEnvelope("uid", mimeTurtle, raw, doc).Bytes()
	require.NoError(t, err)

	env, err := ParseEnvelope(data)
	require.NoError(t, err)
	require.NotNil(t, env.Document)
	assert.Len(t, env.Document.Policies, 3)
}

func TestParseEnvelope_Empty(t *testing.T) {
	t.Parallel()
	_, err := ParseEnvelope([]byte("   \n  "))
	assert.Error(t, err)
}
