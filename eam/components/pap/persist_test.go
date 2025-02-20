package pap

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/storage/valkeyrie/memory"
)

func TestNewStore(t *testing.T) {
	t.Run("new store", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewStore(ctx, client, "/base")
		require.NotNil(t, s)

		p, err := NewPolicyFromData("1", "blanco", "", "", strings.NewReader("no content"))
		require.NoError(t, err)
		require.NotNil(t, p)

		p2, err2 := s.Create(p)
		require.NoError(t, err2)
		require.NotNil(t, p2)
		assert.EqualValues(t, p, p2)

		p3, err3 := s.Read(p.Language(), p.ID())
		require.NoError(t, err3)
		require.NotNil(t, p3)
		assert.EqualValues(t, p, p3)

		p, err = NewPolicyFromData("1", "blanco", "-", "https://localhost:8080/v1/policy/1", strings.NewReader("blanco content"))
		require.NoError(t, err)
		require.NotNil(t, p)

		p4, err4 := s.Update(p3, p)
		require.NoError(t, err4)
		require.NotNil(t, p4)
		assert.EqualValues(t, p, p4)

		p5, err5 := s.Delete(p)
		require.NoError(t, err5)
		require.NotNil(t, p5)
		assert.EqualValues(t, p, p5)
	})
}

func TestNewStore_DupError(t *testing.T) {
	t.Run("new store - duplicate error", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewStore(ctx, client, "/base")
		require.NotNil(t, s)

		p, err := NewPolicyFromData("1", "blanco", "", "", strings.NewReader("no content"))
		require.NoError(t, err)
		require.NotNil(t, p)

		p2, err2 := s.Create(p)
		require.NoError(t, err2)
		require.NotNil(t, p2)
		assert.EqualValues(t, p, p2)

		p3, err3 := s.Create(p)
		require.Error(t, err3)
		require.Nil(t, p3)
	})
}

func TestNewStore_Create_NotFound(t *testing.T) {
	t.Run("new store - create - not found", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewStore(ctx, client, "/base")
		require.NotNil(t, s)

		p2, err2 := s.Read("none", "bad")
		require.Error(t, err2)
		require.Nil(t, p2)
	})
}

func TestNewStore_Update_NotFound(t *testing.T) {
	t.Run("new store - update - not found", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewStore(ctx, client, "/base")
		require.NotNil(t, s)

		p, err := NewPolicyFromData("1", "blanco", "", "", strings.NewReader("no content"))
		require.NoError(t, err)
		require.NotNil(t, p)

		p2, err2 := s.Update(p, p)
		require.Error(t, err2)
		require.Nil(t, p2)
	})
}

func TestNewStore_Delete_NotFound(t *testing.T) {
	t.Run("new store - delete - not found", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewStore(ctx, client, "/base")
		require.NotNil(t, s)

		p, err2 := NewPolicyFromData("x", "y", "", "", strings.NewReader(""))
		require.NoError(t, err2)
		require.NotNil(t, p)

		p2, err3 := s.Delete(p)
		require.Error(t, err3)
		require.Nil(t, p2)
	})
}
