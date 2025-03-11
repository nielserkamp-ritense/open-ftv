package pip

import (
	"context"
	"fmt"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/storage/valkeyrie/memory"
)

func TestMarshalUnmarshalEntities(t *testing.T) {
	testCases := []struct {
		name      string
		a         models.Entity
		wantPanic bool
		wantErr   bool
	}{
		{
			name: "simple",
			a:    models.NewEntity("cedar", "12345", models.NewAttributeSet()),
		},
		{
			name: "with attributes",
			a:    models.NewEntity("cedar", "12345", models.NewAttributeSet(models.NewAttribute("hello", "world"), models.NewAttribute("int", 12345))),
		},
		{
			name: "with parents",
			a:    models.NewEntity("cedar", "12345", models.NewAttributeSet(), "cedar:456", "cedar:789"),
		},
		{
			name: "original with attributes&parents",
			a:    models.NewEntity("cedar", "12345", models.NewAttributeSet(models.NewAttribute("hello", "world"), models.NewAttribute("int", 12345)), "cedar:456", "cedar:789"),
		},
		{
			name:      "panic (1)",
			a:         models.NewEntity("cedar", "12345", models.NewAttributeSet(models.NewAttribute("bad", make(chan byte)))),
			wantPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				e := recover()
				if tc.wantPanic {
					require.NotNil(t, e)
				} else {
					require.Nil(t, e)
				}
			}()

			data := marshalEntity(tc.a)
			require.NotNil(t, data)

			got, err := unmarshalEntity(data)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.True(t, models.EntityEqual(got, tc.a))
			}
		})
	}
}

func TestUnmarshalEntity_Fail(t *testing.T) {
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
			got, err := unmarshalEntity(tc.data)
			require.Error(t, err)
			require.Nil(t, got)
		})
	}
}

func TestNewEntityStore(t *testing.T) {
	t.Run("new entity store", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewEntityStore(ctx, client, "/base")
		require.NotNil(t, s)

		e := models.NewEntity("cedar", "12345", models.NewAttributeSet(models.NewAttribute("hello", "world"), models.NewAttribute("int", 12345)), "cedar:456", "cedar:789")
		require.NotNil(t, e)

		e2, err2 := s.Create(e)
		require.NoError(t, err2)
		require.NotNil(t, e2)
		assert.EqualValues(t, e, e2)

		e3, ix, err3 := s.Read(e.UID())
		require.NoError(t, err3)
		require.NotNil(t, e3)
		assert.EqualValues(t, e, e3)

		e = models.NewEntity("cedar", "12345", models.NewAttributeSet(models.NewAttribute("hello", "world2"), models.NewAttribute("bool", true)), "cedar:654")
		require.NotNil(t, e)

		e4, err4 := s.Update(e3, ix, e)
		require.NoError(t, err4)
		require.NotNil(t, e4)
		assert.EqualValues(t, e, e4)

		e5, err5 := s.Delete(e, ix)
		require.NoError(t, err5)
		require.NotNil(t, e5)
		assert.EqualValues(t, e, e5)
	})
}

func TestNewEntityStore_DupError(t *testing.T) {
	t.Run("new entity store - duplicate error", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewEntityStore(ctx, client, "/base")
		require.NotNil(t, s)

		e := models.NewEntity("cedar", "12345", models.NewAttributeSet(models.NewAttribute("hello", "world"), models.NewAttribute("int", 12345)), "cedar:456", "cedar:789")
		require.NotNil(t, e)

		e2, err2 := s.Create(e)
		require.NoError(t, err2)
		require.NotNil(t, e2)
		assert.EqualValues(t, e, e2)

		e3, err3 := s.Create(e)
		require.Error(t, err3)
		require.Nil(t, e3)
	})
}

func TestNewEntityStore_Read_NotFound(t *testing.T) {
	t.Run("new store - read - not found", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewEntityStore(ctx, client, "/base")
		require.NotNil(t, s)

		e, _, err2 := s.Read("bad")
		require.Error(t, err2)
		require.Nil(t, e)
	})
}

func TestNewEntityStore_Update_NotFound(t *testing.T) {
	t.Run("new entity store - update - not found", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewEntityStore(ctx, client, "/base")
		require.NotNil(t, s)

		e := models.NewEntity("cedar", "12345", models.NewAttributeSet(models.NewAttribute("hello", "world"), models.NewAttribute("int", 12345)), "cedar:456", "cedar:789")
		require.NotNil(t, e)

		e2, err2 := s.Update(e, 0, e)
		require.Error(t, err2)
		require.Nil(t, e2)
	})
}

func TestNewEntityStore_Delete_NotFound(t *testing.T) {
	t.Run("new entity store - delete - not found", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewEntityStore(ctx, client, "/base")
		require.NotNil(t, s)

		e := models.NewEntity("cedar", "12345", models.NewAttributeSet(models.NewAttribute("hello", "world"), models.NewAttribute("int", 12345)), "cedar:456", "cedar:789")
		require.NotNil(t, e)

		e2, err3 := s.Delete(e, 0)
		require.Error(t, err3)
		require.Nil(t, e2)
	})
}

func TestNewEntityStore_List(t *testing.T) {
	t.Run("new entity store - list", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewEntityStore(ctx, client, "/base")
		require.NotNil(t, s)

		got, err := s.List()
		require.Error(t, err)
		require.Nil(t, got)

		for i := range 5 {
			e := models.NewEntity("cedar", fmt.Sprintf("k%d", i+1), models.NewAttributeSet(models.NewAttribute("hello", "world")))
			require.NotNil(t, e)

			_, err = s.Create(e)
			require.NoError(t, err)
		}

		got, err = s.List()
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, 5, len(got))
	})
}
