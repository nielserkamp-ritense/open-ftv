package pap

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestReadByHash_ResolvesExactContent(t *testing.T) {
	t.Parallel()

	client := memory.New()
	defer client.Close()
	s := NewStore(context.Background(), client, "/base")

	const v1 = "policy source v1"
	p1, err := NewPolicyFromData("p", "odrl", "", "", strings.NewReader(v1))
	require.NoError(t, err)
	_, err = s.Create(p1)
	require.NoError(t, err)

	h1 := HashContent([]byte(v1))
	got, err := s.ReadByHash(h1)
	require.NoError(t, err)
	require.NotNil(t, got)
	body, _ := io.ReadAll(got.Content())
	assert.Equal(t, v1, string(body))

	// update to v2: v1 must remain retrievable under its own hash (retention).
	_, idx, err := s.Read("odrl", "p")
	require.NoError(t, err)
	const v2 = "policy source v2 mutated"
	p2, err := NewPolicyFromData("p", "odrl", "", "", strings.NewReader(v2))
	require.NoError(t, err)
	_, err = s.Update(p1, idx, p2)
	require.NoError(t, err)

	old, err := s.ReadByHash(h1)
	require.NoError(t, err)
	oldBody, _ := io.ReadAll(old.Content())
	assert.Equal(t, v1, string(oldBody), "old version must remain resolvable by its hash")

	cur, err := s.ReadByHash(HashContent([]byte(v2)))
	require.NoError(t, err)
	curBody, _ := io.ReadAll(cur.Content())
	assert.Equal(t, v2, string(curBody))

	// unknown hash fails.
	_, err = s.ReadByHash("deadbeef")
	require.Error(t, err)
}
