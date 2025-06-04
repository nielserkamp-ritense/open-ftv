package csv

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

func TestEncoder_WriteSeparator(t *testing.T) {
	t.Parallel()

	t.Run("separator", func(t *testing.T) {
		t.Parallel()

		enc := NewBytesEncoder()
		require.NotNil(t, enc)

		enc2, ok := enc.(*encoder)
		require.True(t, ok)
		require.NotNil(t, enc2)

		enc2.WriteSeparator()

		got := string(enc.Bytes())
		assert.Equal(t, ",", got)
	})
}

func TestEncoder_WriteEOL(t *testing.T) {
	t.Parallel()

	t.Run("eol", func(t *testing.T) {
		t.Parallel()

		enc := NewBytesEncoder()
		require.NotNil(t, enc)

		enc2, ok := enc.(*encoder)
		require.True(t, ok)
		require.NotNil(t, enc2)

		enc2.WriteSeparator()
		enc2.WriteEOL()

		got := string(enc.Bytes())
		assert.Equal(t, "\n", got)
	})
}

func TestEncoder_Encode(t *testing.T) {
	t.Parallel()

	def := &schema.Field{
		Object: schema.Object{
			Parent: schema.Parent{ID: "x"},
			Fields: []*schema.Field{
				{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: enums.StringType},
				{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}, Type: enums.StringType},
				{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}, Type: enums.IntegerType},
				{Object: schema.Object{Parent: schema.Parent{ID: "f4"}}, Type: enums.BooleanType},
			},
		},
		Type: enums.ObjectType,
	}

	testCases := []struct {
		name string
		def  *schema.Field
		in   any
		want string
	}{
		{name: "string", def: &schema.Field{Type: enums.StringType}, in: "hello world", want: `"hello world"`},
		{name: "integer", def: &schema.Field{Type: enums.IntegerType}, in: int64(123456789), want: "123456789"},
		{name: "unsigned integer", def: &schema.Field{Type: enums.UnsignedIntegerType}, in: uint64(987654321), want: "987654321"},
		{name: "float", def: &schema.Field{Type: enums.FloatType}, in: 5.6789, want: "5.6789"},
		{name: "bool", def: &schema.Field{Type: enums.BooleanType}, in: true, want: `"true"`},
		{name: "date", def: &schema.Field{Type: enums.DateType, Format: "brp-date"}, in: time.Date(2025, 11, 12, 13, 14, 15, 0, time.UTC), want: `"20251112"`},
		{name: "record", def: def, in: map[string]any{"f2": "yo yo", "f4": 1, "f1": 123, "f3": "987"}, want: `"{\"f1\":\"123\",\"f2\":\"yo yo\",\"f3\":987,\"f4\":true}"`},
		{name: "strings", def: &schema.Field{Type: enums.StringType, IsArray: true}, in: []string{"hello", "world"}, want: `"hello,world"`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			enc := NewBytesEncoder()
			require.NotNil(t, enc)

			enc2, ok := enc.(*encoder)
			require.True(t, ok)
			require.NotNil(t, enc2)

			enc2.Encode(tc.def, tc.in)

			got := string(enc.Bytes())
			assert.Equal(t, tc.want, got)
		})
	}
}
