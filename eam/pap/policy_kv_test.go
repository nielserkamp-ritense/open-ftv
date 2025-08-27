package pap

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestNewStore(t *testing.T) {
	t.Parallel()

	t.Run("new store", func(t *testing.T) {
		t.Parallel()

		client := memory.New()
		require.NotNil(t, client)
		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewKeyValueDB(client, "/base")
		require.NotNil(t, s)

		p, err := models.NewPolicyFromData("1", "blanco", "", "", strings.NewReader("no content"))
		require.NoError(t, err)
		require.NotNil(t, p)

		p2, err2 := s.CreatePolicy(ctx, p)
		require.NoError(t, err2)
		require.NotNil(t, p2)
		assert.EqualValues(t, p, p2)

		p3, _, err3 := s.ReadPolicy(ctx, p.ID())
		require.NoError(t, err3)
		require.NotNil(t, p3)
		assert.EqualValues(t, p, p3)

		p, err = models.NewPolicyFromData("1", "blanco", "-", "https://localhost:8080/v1/policy/1", strings.NewReader("blanco content"))
		require.NoError(t, err)
		require.NotNil(t, p)

		p4, err4 := s.UpdatePolicy(ctx, p3, 0, p)
		require.NoError(t, err4)
		require.NotNil(t, p4)
		assert.EqualValues(t, p, p4)

		p5, err5 := s.DeletePolicy(ctx, p, 0)
		require.NoError(t, err5)
		require.NotNil(t, p5)
		assert.EqualValues(t, p, p5)
	})
}

func TestNewStore_DupError(t *testing.T) {
	t.Parallel()

	t.Run("new store - duplicate error", func(t *testing.T) {
		t.Parallel()

		client := memory.New()
		require.NotNil(t, client)
		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewKeyValueDB(client, "/base")
		require.NotNil(t, s)

		p, err := models.NewPolicyFromData("1", "blanco", "", "", strings.NewReader("no content"))
		require.NoError(t, err)
		require.NotNil(t, p)

		p2, err2 := s.CreatePolicy(ctx, p)
		require.NoError(t, err2)
		require.NotNil(t, p2)
		assert.EqualValues(t, p, p2)

		p3, err3 := s.CreatePolicy(ctx, p)
		require.Error(t, err3)
		require.Nil(t, p3)
	})
}

func TestNewStore_Read_NotFound(t *testing.T) {
	t.Parallel()

	t.Run("new store - read - not found", func(t *testing.T) {
		t.Parallel()

		client := memory.New()
		require.NotNil(t, client)
		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewKeyValueDB(client, "/base")
		require.NotNil(t, s)

		p2, _, err2 := s.ReadPolicy(ctx, "bad")
		require.NoError(t, err2)
		require.Nil(t, p2)
	})
}

func TestNewStore_Update_NotFound(t *testing.T) {
	t.Parallel()

	t.Run("new store - update - not found", func(t *testing.T) {
		t.Parallel()

		client := memory.New()
		require.NotNil(t, client)
		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewKeyValueDB(client, "/base")
		require.NotNil(t, s)

		p, err := models.NewPolicyFromData("1", "blanco", "", "", strings.NewReader("no content"))
		require.NoError(t, err)
		require.NotNil(t, p)

		p2, err2 := s.UpdatePolicy(ctx, p, 0, p)
		require.Error(t, err2)
		require.Nil(t, p2)
	})
}

func TestNewStore_Delete_NotFound(t *testing.T) {
	t.Parallel()

	t.Run("new store - delete - not found", func(t *testing.T) {
		t.Parallel()

		client := memory.New()
		require.NotNil(t, client)
		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewKeyValueDB(client, "/base")
		require.NotNil(t, s)

		p, err2 := models.NewPolicyFromData("x", "y", "", "", strings.NewReader(""))
		require.NoError(t, err2)
		require.NotNil(t, p)

		p2, err3 := s.DeletePolicy(ctx, p, 0)
		require.Error(t, err3)
		require.Nil(t, p2)
	})
}

func TestNewStore_List(t *testing.T) {
	t.Parallel()

	client := memory.New()
	require.NotNil(t, client)
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := NewKeyValueDB(client, "/base")
	require.NotNil(t, s)

	policies := []struct {
		language string
		id       string
	}{
		{language: "english", id: "1"},
		{language: "english", id: "2"},
		{language: "english", id: "3"},
		{language: "dutch", id: "4"},
		{language: "dutch", id: "5"},
		{language: "french", id: "6"},
	}

	for _, data := range policies {
		p, err := models.NewPolicyFromData(data.id, data.language, "", "", strings.NewReader("no content"))
		require.NoError(t, err)
		require.NotNil(t, p)

		p2, err2 := s.CreatePolicy(ctx, p)
		require.NoError(t, err2)
		require.NotNil(t, p2)
		assert.EqualValues(t, p, p2)
	}

	testCases := []struct {
		name      string
		language  string
		wantCount int
	}{
		{
			name:      "no language",
			wantCount: 6,
		},
		{
			name:      "english",
			language:  "english",
			wantCount: 3,
		},
		{
			name:      "dutch",
			language:  "dutch",
			wantCount: 2,
		},
		{
			name:     "bad language",
			language: "oops",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			list, err4 := s.ListPolicies(ctx, tc.language)
			require.NoError(t, err4)
			require.NotNil(t, list)
			assert.EqualValues(t, tc.wantCount, len(list))
		})
	}
}

func TestRead_UnmarshalError(t *testing.T) {
	t.Parallel()

	t.Run("read - unmarshal error", func(t *testing.T) {
		t.Parallel()

		client := memory.New()
		require.NotNil(t, client)
		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewKeyValueDB(client, "/base")
		require.NotNil(t, s)

		p, err := models.NewPolicyFromData("1", "blanco", "", "", strings.NewReader("no content"))
		require.NoError(t, err)
		require.NotNil(t, p)

		p2, err2 := s.CreatePolicy(ctx, p)
		require.NoError(t, err2)
		require.NotNil(t, p2)
		assert.EqualValues(t, p, p2)

		err3 := client.Put(ctx, "/base/1", []byte("\000\001"), writeOptions)
		require.NoError(t, err3)

		p3, _, err4 := s.ReadPolicy(ctx, p.ID())
		require.Error(t, err4)
		require.Nil(t, p3)

		list, err5 := s.ListPolicies(ctx, "")
		require.Error(t, err5)
		require.Nil(t, list)
	})
}
