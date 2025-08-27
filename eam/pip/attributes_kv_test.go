package pip

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestMarshalUnmarshalAttributes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		a         *models.Attribute
		wantPanic bool
		wantErr   bool
	}{
		{
			name: "simple",
			a:    models.NewAttribute("k1", 123),
		},
		{
			name: "with type",
			a:    models.NewAttributeWithType("k1", 123, "xsd:integer"),
		},
		{
			name: "original",
			a:    models.NewOriginalAttribute("k1", 123, "123", ""),
		},
		{
			name: "original with type",
			a:    models.NewOriginalAttribute("k1", 123, "123", "xsd:positiveNumber"),
		},
		{
			name:      "panic (1)",
			a:         models.NewAttribute("k1", make(chan byte)),
			wantPanic: true,
		},
		{
			name:      "panic (2)",
			a:         models.NewOriginalAttribute("k1", 123, make(chan byte), ""),
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

			data := marshalAttribute(tc.a)
			require.NotNil(t, data)

			got, err := unmarshalAttribute(data)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.True(t, got.Equals(tc.a))
			}
		})
	}
}

func TestMarshalAttribute_Repeated(t *testing.T) {
	t.Parallel()

	t.Run("marshal repeated", func(t *testing.T) {
		a := models.NewOriginalAttribute("k1", 123, "123", "xsd:positiveNumber")

		b1 := marshalAttribute(a)

		for i := range 100 {
			b2 := marshalAttribute(a)
			assert.EqualValuesf(t, b1, b2, fmt.Sprintf("repeated entity %d", i))
		}
	})
}

func TestUnmarshalAttribute_Fail(t *testing.T) {
	t.Parallel()

	d1, _ := json.Marshal(&attribute{
		Key:   "k1",
		Value: "this is not base64! \000\001",
	})

	d2, _ := json.Marshal(&attribute{
		Key:   "k2",
		Value: "aGFoYQ==",
	})

	var q any = "hello world"
	b := &bytes.Buffer{}
	err := gob.NewEncoder(b).Encode(&q)
	require.NoError(t, err)

	b64 := base64.StdEncoding.EncodeToString(b.Bytes())

	d3, _ := json.Marshal(&attribute{
		Key:      "k3",
		Value:    b64,
		Original: "haha!\000",
	})

	d4, _ := json.Marshal(&attribute{
		Key:      "k3",
		Value:    b64,
		Original: "aGFoYQ==",
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
			name: "value not base64",
			data: d1,
		},
		{
			name: "value nog gob",
			data: d2,
		},
		{
			name: "original not base64",
			data: d3,
		},
		{
			name: "original not gob",
			data: d4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := unmarshalAttribute(tc.data)
			require.Error(t, err)
			require.Nil(t, got)
		})
	}
}

func TestNewAttributeStore(t *testing.T) {
	t.Parallel()

	t.Run("new attribute store", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewAttributeStore(client, "/base")
		require.NotNil(t, s)

		a := models.NewOriginalAttribute("k1", 123, "123", "xsd:integer")
		require.NotNil(t, a)

		a2, err2 := s.CreateAttribute(ctx, a)
		require.NoError(t, err2)
		require.NotNil(t, a2)
		assert.EqualValues(t, a, a2)

		a3, ix, err3 := s.ReadAttribute(ctx, a.Key())
		require.NoError(t, err3)
		require.NotNil(t, a3)
		assert.EqualValues(t, a, a3)

		a = models.NewOriginalAttribute("k1", "haha", "haha", "xsd:string")
		require.NotNil(t, a)

		a4, err4 := s.UpdateAttribute(ctx, a3, ix, a)
		require.NoError(t, err4)
		require.NotNil(t, a4)
		assert.EqualValues(t, a, a4)

		a5, err5 := s.DeleteAttribute(ctx, a, ix)
		require.NoError(t, err5)
		require.NotNil(t, a5)
		assert.EqualValues(t, a, a5)
	})
}

func TestNewAttributeStore_DupError(t *testing.T) {
	t.Parallel()

	t.Run("new attribute store - duplicate error", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewAttributeStore(client, "/base")
		require.NotNil(t, s)

		a := models.NewOriginalAttribute("k1", 123, "123", "")
		require.NotNil(t, a)

		a2, err2 := s.CreateAttribute(ctx, a)
		require.NoError(t, err2)
		require.NotNil(t, a2)
		assert.EqualValues(t, a, a2)

		a3, err3 := s.CreateAttribute(ctx, a)
		require.Error(t, err3)
		require.Nil(t, a3)
	})
}

func TestNewAttributeStore_Read_NotFound(t *testing.T) {
	t.Parallel()

	t.Run("new store - read - not found", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewAttributeStore(client, "/base")
		require.NotNil(t, s)

		a2, _, err2 := s.ReadAttribute(ctx, "bad")
		require.NoError(t, err2)
		require.Nil(t, a2)
	})
}

func TestNewAttributeStore_Update_NotFound(t *testing.T) {
	t.Parallel()

	t.Run("new attribute store - update - not found", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewAttributeStore(client, "/base")
		require.NotNil(t, s)

		a := models.NewOriginalAttribute("k1", 123, "123", "xsd:positiveNumber")
		require.NotNil(t, a)

		a2, err2 := s.UpdateAttribute(ctx, a, 0, a)
		require.Error(t, err2)
		require.Nil(t, a2)
	})
}

func TestNewAttributeStore_Delete_NotFound(t *testing.T) {
	t.Parallel()

	t.Run("new attribute store - delete - not found", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewAttributeStore(client, "/base")
		require.NotNil(t, s)

		a := models.NewOriginalAttribute("k1", 123, "123", "xsd:positiveNumber")
		require.NotNil(t, a)

		a2, err3 := s.DeleteAttribute(ctx, a, 0)
		require.Error(t, err3)
		require.Nil(t, a2)
	})
}

func TestNewAttributeStore_List(t *testing.T) {
	t.Parallel()

	t.Run("new attribute store - list", func(t *testing.T) {
		client := memory.New()
		require.NotNil(t, client)

		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := NewAttributeStore(client, "/base")
		require.NotNil(t, s)

		got, err := s.ListAttributes(ctx)
		require.Error(t, err)
		require.Nil(t, got)

		for i := range 5 {
			a := models.NewOriginalAttribute(fmt.Sprintf("k%d", i+1), 123, "123", "xsd:positiveNumber")
			require.NotNil(t, a)

			_, err = s.CreateAttribute(ctx, a)
			require.NoError(t, err)
		}

		got, err = s.ListAttributes(ctx)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, 5, len(got))
	})
}
