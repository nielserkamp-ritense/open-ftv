package pip

import (
	"context"
	"fmt"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestMarshalUnmarshalEntities(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		e         *models.Entity
		wantPanic bool
		wantErr   bool
	}{
		{
			name: "simple",
			e:    models.NewEntity("cedar", "12345", models.NewAttributeSet()),
		},
		{
			name: "with attributes",
			e:    models.NewEntity("cedar", "12345", models.NewAttributeSet(models.NewAttribute("hello", "world"), models.NewAttribute("int", 12345))),
		},
		{
			name: "with parents",
			e:    models.NewEntity("cedar", "12345", models.NewAttributeSet(), "cedar:456", "cedar:789"),
		},
		{
			name: "original with attributes&parents",
			e:    models.NewEntity("cedar", "12345", models.NewAttributeSet(models.NewAttribute("hello", "world"), models.NewAttribute("int", 12345)), "cedar:456", "cedar:789"),
		},
		{
			name:      "panic (1)",
			e:         models.NewEntity("cedar", "12345", models.NewAttributeSet(models.NewAttribute("bad", make(chan byte)))),
			wantPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				e := recover()
				if tc.wantPanic {
					require.NotNil(t, e)
				} else {
					require.Nil(t, e)
				}
			}()

			data := marshalEntity(tc.e)
			require.NotNil(t, data)

			got, err := unmarshalEntity(data)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.True(t, got.Equals(tc.e))
			}
		})
	}
}

func TestMarshalEntity_Repeated(t *testing.T) {
	t.Parallel()

	t.Run("marshal repeated", func(t *testing.T) {
		e := models.NewEntity("cedar", "12345", models.NewAttributeSet(models.NewAttribute("hello", "world"), models.NewAttribute("int", 12345)), "cedar:456", "cedar:789")

		b1 := marshalEntity(e)

		for i := range 100 {
			b2 := marshalEntity(e)
			assert.EqualValuesf(t, b1, b2, fmt.Sprintf("repeated entity %d", i))
		}
	})
}

func TestUnmarshalEntity_Fail(t *testing.T) {
	t.Parallel()

	d1, _ := json.Marshal(&entity{
		Type:       "cedar",
		ID:         "12345",
		Attributes: "this is not base64! \000\001",
	})

	d2, _ := json.Marshal(&entity{
		Type:       "cedar",
		ID:         "12345",
		Attributes: "aGFoYQ==",
	})

	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "bad input",
			data: []byte{'[', '\000', '!'},
		},
		{
			name: "attributes not base64",
			data: d1,
		},
		{
			name: "attributes nog gob",
			data: d2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := unmarshalEntity(tc.data)
			require.Error(t, err)
			require.Nil(t, got)
		})
	}
}

func TestNewEntityStore(t *testing.T) {
	t.Parallel()

	t.Run("new entity store", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewEntityStore(client, "/base")
		require.NotNil(t, s)

		e := models.NewEntity("cedar", "12345", models.NewAttributeSet(models.NewAttribute("hello", "world"), models.NewAttribute("int", 12345)), "cedar:456", "cedar:789")
		require.NotNil(t, e)

		e2, err2 := s.CreateEntity(ctx, identity.NewSystemPrincipal(), e)
		require.NoError(t, err2)
		require.NotNil(t, e2)
		assert.EqualValues(t, e, e2)

		e3, ix, err3 := s.ReadEntity(ctx, e.Type(), e.ID())
		require.NoError(t, err3)
		require.NotNil(t, e3)
		assert.EqualValues(t, e, e3)

		e4 := models.NewEntity(e3.Type(), e3.ID(), models.NewAttributeSet(models.NewAttribute("hello", "world2"), models.NewAttribute("bool", true)), "cedar:654")
		require.NotNil(t, e4)

		e5, err4 := s.UpdateEntity(ctx, identity.NewSystemPrincipal(), e3, ix, e4)
		require.NoError(t, err4)
		require.NotNil(t, e5)
		assert.EqualValues(t, e4, e5)

		e6, err5 := s.DeleteEntity(ctx, identity.NewSystemPrincipal(), e4, ix)
		require.NoError(t, err5)
		require.NotNil(t, e6)
		assert.EqualValues(t, e4, e6)
	})
}

func TestNewEntityStore_DupError(t *testing.T) {
	t.Parallel()

	t.Run("new entity store - duplicate error", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewEntityStore(client, "/base")
		require.NotNil(t, s)

		e := models.NewEntity("cedar", "12345", models.NewAttributeSet(models.NewAttribute("hello", "world"), models.NewAttribute("int", 12345)), "cedar:456", "cedar:789")
		require.NotNil(t, e)

		e2, err2 := s.CreateEntity(ctx, identity.NewSystemPrincipal(), e)
		require.NoError(t, err2)
		require.NotNil(t, e2)
		assert.EqualValues(t, e, e2)

		e3, err3 := s.CreateEntity(ctx, identity.NewSystemPrincipal(), e)
		require.Error(t, err3)
		require.Nil(t, e3)
	})
}

func TestNewEntityStore_Read_NotFound(t *testing.T) {
	t.Parallel()

	t.Run("new store - read - not found", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewEntityStore(client, "/base")
		require.NotNil(t, s)

		e, _, err2 := s.ReadEntity(ctx, "q", "bad")
		require.NoError(t, err2)
		require.Nil(t, e)
	})
}

func TestNewEntityStore_Update_NotFound(t *testing.T) {
	t.Parallel()

	t.Run("new entity store - update - not found", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewEntityStore(client, "/base")
		require.NotNil(t, s)

		e := models.NewEntity("cedar", "12345", models.NewAttributeSet(models.NewAttribute("hello", "world"), models.NewAttribute("int", 12345)), "cedar:456", "cedar:789")
		require.NotNil(t, e)

		e2, err2 := s.UpdateEntity(ctx, identity.NewSystemPrincipal(), e, 0, e)
		require.Error(t, err2)
		require.Nil(t, e2)
	})
}

func TestNewEntityStore_Delete_NotFound(t *testing.T) {
	t.Parallel()

	t.Run("new entity store - delete - not found", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewEntityStore(client, "/base")
		require.NotNil(t, s)

		e := models.NewEntity("cedar", "12345", models.NewAttributeSet(models.NewAttribute("hello", "world"), models.NewAttribute("int", 12345)), "cedar:456", "cedar:789")
		require.NotNil(t, e)

		e2, err3 := s.DeleteEntity(ctx, identity.NewSystemPrincipal(), e, 0)
		require.Error(t, err3)
		require.Nil(t, e2)
	})
}

func TestNewEntityStore_List(t *testing.T) {
	t.Parallel()

	t.Run("new entity store - list", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewEntityStore(client, "/base")
		require.NotNil(t, s)

		got, err := s.ListEntities(ctx)
		require.Error(t, err)
		require.Nil(t, got)

		for i := range 5 {
			e := models.NewEntity("cedar", fmt.Sprintf("k%d", i+1), models.NewAttributeSet(models.NewAttribute("hello", "world")))
			require.NotNil(t, e)

			_, err = s.CreateEntity(ctx, identity.NewSystemPrincipal(), e)
			require.NoError(t, err)
		}

		got, err = s.ListEntities(ctx)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, 5, len(got))
	})
}
