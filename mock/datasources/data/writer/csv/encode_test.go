package csv

import (
	"strconv"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/types"
)

func TestEncodeValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		def  *schema.Field
		in   any
		want string
	}{
		{name: "string", def: &schema.Field{Type: types.StringType}, in: "hello world", want: `"hello world"`},
		{name: "url", def: &schema.Field{Type: types.URLType}, in: "http://localhost:8080/v1/dataspace", want: `"http://localhost:8080/v1/dataspace"`},
		{name: "email", def: &schema.Field{Type: types.EmailType}, in: "donald@disney.com", want: `"donald@disney.com"`},
		{name: "phone", def: &schema.Field{Type: types.PhoneNrType}, in: "112", want: `"112"`},
		{name: "ip address", def: &schema.Field{Type: types.IPAddressType}, in: "127.0.0.1", want: `"127.0.0.1"`},
		{name: "integer", def: &schema.Field{Type: types.IntegerType}, in: int64(123456789), want: `"123456789"`},
		{name: "unsigned integer", def: &schema.Field{Type: types.UnsignedIntegerType}, in: uint64(987654321), want: `"987654321"`},
		{name: "float", def: &schema.Field{Type: types.FloatType}, in: 5.6789, want: `"5.6789"`},
		{name: "bool", def: &schema.Field{Type: types.BooleanType}, in: true, want: `"true"`},
		// {name: "date", def: &schema.Field{Type: types.DateType, Format: "brp-date"}, in: time.Date(2025, 11, 12, 13, 14, 15, 0, time.UTC), want: `"20251112"`},
		// {name: "time", def: &schema.Field{Type: types.TimeType}, in: time.Date(0, 1, 1, 13, 14, 15, 0, time.UTC), want: `"13:14:15"`},
		// {name: "datetime", def: &schema.Field{Type: types.DateTimeType}, in: time.Date(2025, 11, 12, 13, 14, 15, 999000000, time.UTC), want: `"2025-11-12 13:14:15.999"`},
		// {name: "object", def:  &schema.Field{Type: types.ObjectType}, in:   nil, want: ""},
		{name: "invalid type", def: &schema.Field{Type: 244}, in: &struct{}{}, want: `"&{}"`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			v := encodeValues(tc.def, tc.in)

			enc := NewBytesEncoder()
			require.NotNil(t, enc)

			enc2, ok := enc.(*encoder)
			require.True(t, ok)
			require.NotNil(t, enc2)

			switch tp := v.(type) {
			case string:
				enc2.WriteString(tp)
			case int64:
				enc2.Buffer.WriteString(strconv.FormatInt(tp, 10))
			case uint64:
				enc2.Buffer.WriteString(strconv.FormatUint(tp, 10))
			case float64:
				enc2.Buffer.WriteString(strconv.FormatFloat(tp, 'g', -1, 64))
			case bool:
				enc2.WriteString(encodeBool(tp))
			case time.Time:
				enc2.WriteString(encodeTime(tc.def, tp))
			default:
				b, _ := json.Marshal(v)
				enc2.WriteString(string(b))
			}

			got := string(enc.Bytes())
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestEncodeValue(t *testing.T) {
	t.Parallel()

	def := &schema.Field{
		Object: schema.Object{
			Parent: schema.Parent{ID: "x"},
			Fields: []*schema.Field{
				{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: types.StringType},
				{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}, Type: types.StringType},
				{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}, Type: types.IntegerType},
				{Object: schema.Object{Parent: schema.Parent{ID: "f4"}}, Type: types.BooleanType},
			},
		},
		Type:    types.ObjectType,
		IsArray: true,
	}

	testCases := []struct {
		name string
		def  *schema.Field
		in   any
		want string
	}{
		{name: "string", def: &schema.Field{Type: types.StringType}, in: "hello world", want: `"hello world"`},
		{name: "url", def: &schema.Field{Type: types.URLType}, in: "http://localhost:8080/v1/dataspace", want: `"http://localhost:8080/v1/dataspace"`},
		{name: "email", def: &schema.Field{Type: types.EmailType}, in: "donald@disney.com", want: `"donald@disney.com"`},
		{name: "phone", def: &schema.Field{Type: types.PhoneNrType}, in: "112", want: `"112"`},
		{name: "ip address", def: &schema.Field{Type: types.IPAddressType}, in: "127.0.0.1", want: `"127.0.0.1"`},
		{name: "integer", def: &schema.Field{Type: types.IntegerType}, in: int64(123456789), want: "123456789"},
		{name: "unsigned integer", def: &schema.Field{Type: types.UnsignedIntegerType}, in: uint64(987654321), want: "987654321"},
		{name: "float", def: &schema.Field{Type: types.FloatType}, in: 5.6789, want: "5.6789"},
		{name: "bool", def: &schema.Field{Type: types.BooleanType}, in: true, want: `"true"`},
		{name: "date", def: &schema.Field{Type: types.DateType, Format: "brp-date"}, in: time.Date(2025, 11, 12, 13, 14, 15, 0, time.UTC), want: `"20251112"`},
		{name: "time", def: &schema.Field{Type: types.TimeType}, in: time.Date(0, 1, 1, 13, 14, 15, 0, time.UTC), want: `"13:14:15"`},
		{name: "datetime", def: &schema.Field{Type: types.DateTimeType}, in: time.Date(2025, 11, 12, 13, 14, 15, 999000000, time.UTC), want: `"2025-11-12 13:14:15.999"`},
		{name: "object", def: def, in: map[string]any{"f4": false, "f1": "yo", "f2": 123, "f3": 159}, want: `"{\"f1\":\"yo\",\"f2\":\"123\",\"f3\":159,\"f4\":false}"`},
		{name: "invalid type", def: &schema.Field{Type: 244}, in: &struct{}{}, want: `"&{}"`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			v := encodeValue(tc.def, tc.in)

			enc := NewBytesEncoder()
			require.NotNil(t, enc)

			enc2, ok := enc.(*encoder)
			require.True(t, ok)
			require.NotNil(t, enc2)

			switch tp := v.(type) {
			case string:
				enc2.WriteString(tp)
			case int64:
				enc2.Buffer.WriteString(strconv.FormatInt(tp, 10))
			case uint64:
				enc2.Buffer.WriteString(strconv.FormatUint(tp, 10))
			case float64:
				enc2.Buffer.WriteString(strconv.FormatFloat(tp, 'g', -1, 64))
			case bool:
				enc2.WriteString(encodeBool(tp))
			case time.Time:
				enc2.WriteString(encodeTime(tc.def, tp))
			default:
				b, _ := json.Marshal(v)
				enc2.WriteString(string(b))
			}

			got := string(enc.Bytes())
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestEncodeStrings(t *testing.T) {
	t.Parallel()

	t.Run("strings", func(t *testing.T) {
		t.Parallel()

		rec := []any{true, "hello world", 987.6543, "yo"}

		got := encodeStrings(nil, rec)
		want := "true,hello world,987.6543,yo"
		assert.Equal(t, want, got)
	})
}

func TestEncodeObjects(t *testing.T) {
	t.Parallel()

	def := &schema.Field{
		Object: schema.Object{
			Parent: schema.Parent{ID: "x"},
			Fields: []*schema.Field{
				{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: types.StringType},
				{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}, Type: types.StringType},
				{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}, Type: types.IntegerType},
				{Object: schema.Object{Parent: schema.Parent{ID: "f4"}}, Type: types.BooleanType},
			},
		},
		Type:    types.ObjectType,
		IsArray: true,
	}

	testCases := []struct {
		name string
		in   any
		want any
	}{
		{
			name: "any slice",
			in:   []any{[]any{true, "hello world", 987.6543, "true"}},
			want: []any{map[string]any{"f1": "true", "f2": "hello world", "f3": int64(988), "f4": true}},
		},
		{
			name: "maps",
			in:   []map[string]any{{"f4": false, "f1": "yo", "f2": 123, "f3": 159}},
			want: []map[string]any{{"f1": "yo", "f2": "123", "f3": int64(159), "f4": false}},
		},
		{
			name: "slices",
			in: [][]any{
				{"jupiter", "big", 555, true},
				{"pluto", "pretty big", 777, false},
			},
			want: []map[string]interface{}{
				{"f1": "jupiter", "f2": "big", "f3": int64(555), "f4": true},
				{"f1": "pluto", "f2": "pretty big", "f3": int64(777), "f4": false},
			},
		},
		{
			name: "unsupported",
			in:   &struct{}{},
			want: "&{}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := encodeObjects(def, tc.in)
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestEncodeAnySlice(t *testing.T) {
	t.Parallel()

	t.Run("any slice", func(t *testing.T) {
		t.Parallel()

		def := &schema.Field{
			Object: schema.Object{
				Parent: schema.Parent{ID: "x"},
				Fields: []*schema.Field{
					{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: types.StringType},
					{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}, Type: types.StringType},
					{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}, Type: types.IntegerType},
					{Object: schema.Object{Parent: schema.Parent{ID: "f4"}}, Type: types.BooleanType},
				},
			},
			Type:    types.ObjectType,
			IsArray: true,
		}

		recs := []any{
			map[string]any{"f1": "hello", "f2": "world", "f4": true, "f3": 123},
			[]any{"goodbye", "world", 321, false},
		}

		got := encodeAnySlice(def, recs)
		want := []any{
			map[string]any{"f1": "hello", "f2": "world", "f3": int64(123), "f4": true},
			map[string]any{"f1": "goodbye", "f2": "world", "f3": int64(321), "f4": false},
		}
		assert.EqualValues(t, want, got)
	})
}

func TestEncodeMaps(t *testing.T) {
	t.Parallel()

	t.Run("maps", func(t *testing.T) {
		t.Parallel()

		def := &schema.Field{
			Object: schema.Object{
				Parent: schema.Parent{ID: "x"},
				Fields: []*schema.Field{
					{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: types.StringType},
					{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}, Type: types.StringType},
					{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}, Type: types.IntegerType},
					{Object: schema.Object{Parent: schema.Parent{ID: "f4"}}, Type: types.BooleanType},
				},
			},
			Type:    types.ObjectType,
			IsArray: true,
		}

		recs := []map[string]any{
			{"f1": "hello", "f2": "world", "f4": true, "f3": 123},
			{"f1": "goodbye", "f2": "world", "f4": false, "f3": 321},
		}

		got := encodeMaps(def, recs)
		want := []map[string]any{
			{"f1": "hello", "f2": "world", "f3": int64(123), "f4": true},
			{"f1": "goodbye", "f2": "world", "f3": int64(321), "f4": false},
		}
		assert.EqualValues(t, want, got)
	})
}

func TestEncodeSlices(t *testing.T) {
	t.Parallel()

	t.Run("slices", func(t *testing.T) {
		t.Parallel()

		def := &schema.Field{
			Object: schema.Object{
				Parent: schema.Parent{ID: "x"},
				Fields: []*schema.Field{
					{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: types.StringType},
					{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}, Type: types.StringType},
					{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}, Type: types.IntegerType},
					{Object: schema.Object{Parent: schema.Parent{ID: "f4"}}, Type: types.BooleanType},
				},
			},
			Type:    types.ObjectType,
			IsArray: true,
		}

		recs := [][]any{
			{"hello", "world", 123, true},
			{"goodbye", "world", 321, false},
		}

		got := encodeSlices(def, recs)
		want := []map[string]any{
			{"f1": "hello", "f2": "world", "f3": int64(123), "f4": true},
			{"f1": "goodbye", "f2": "world", "f3": int64(321), "f4": false},
		}
		assert.EqualValues(t, want, got)
	})
}

func TestEncodeObject(t *testing.T) {
	t.Parallel()

	def := &schema.Field{
		Object: schema.Object{
			Parent: schema.Parent{ID: "x"},
			Fields: []*schema.Field{
				{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: types.IntegerType},
				{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}, Type: types.BooleanType},
				{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}, Type: types.StringType},
				{Object: schema.Object{Parent: schema.Parent{ID: "f4"}}, Type: types.FloatType},
			},
		},
		Type: types.ObjectType,
	}

	testCases := []struct {
		name string
		def  *schema.Field
		in   any
		want any
	}{
		{name: "nil", def: def, want: ""},
		{name: "encoder", in: &testEncoder{}, want: "hello world"},
		{name: "any", in: &struct{}{}, want: "&{}"},
		{
			name: "map",
			def:  def,
			in:   map[string]any{"f1": 1234, "f2": true, "f3": "hello world", "f4": 987.65, "f5": "oops"},
			want: map[string]any{
				"f1": int64(1234),
				"f2": true,
				"f3": "hello world",
				"f4": 987.65,
			},
		},
		{
			name: "slice",
			def:  def,
			in:   []any{1234, true, "hello world", 987.65, "oops"},
			want: map[string]any{
				"f1": int64(1234),
				"f2": true,
				"f3": "hello world",
				"f4": 987.65,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := encodeObject(tc.def, tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestEncodeMap(t *testing.T) {
	t.Parallel()

	t.Run("map", func(t *testing.T) {
		t.Parallel()

		rec := map[string]any{"f1": 1234, "f2": true, "f3": "hello world", "f4": 987.65}
		def := &schema.Field{
			Object: schema.Object{
				Parent: schema.Parent{ID: "x"},
				Fields: []*schema.Field{
					{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: types.IntegerType},
					{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}, Type: types.BooleanType},
					{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}, Type: types.StringType},
					{Object: schema.Object{Parent: schema.Parent{ID: "f4"}}, Type: types.FloatType},
				},
			},
			Type: types.ObjectType,
		}

		got := encodeMap(def, rec)
		want := map[string]any{
			"f1": int64(1234),
			"f2": true,
			"f3": "hello world",
			"f4": 987.65,
		}
		assert.EqualValues(t, want, got)
	})
}

func TestEncodeSlice(t *testing.T) {
	t.Parallel()

	t.Run("slice", func(t *testing.T) {
		t.Parallel()

		rec := []any{1234, true, "hello world", 987.65}
		def := &schema.Field{
			Object: schema.Object{
				Parent: schema.Parent{ID: "x"},
				Fields: []*schema.Field{
					{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: types.IntegerType},
					{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}, Type: types.BooleanType},
					{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}, Type: types.StringType},
					{Object: schema.Object{Parent: schema.Parent{ID: "f4"}}, Type: types.FloatType},
				},
			},
			Type: types.ObjectType,
		}

		got := encodeSlice(def, rec)
		want := map[string]any{
			"f1": int64(1234),
			"f2": true,
			"f3": "hello world",
			"f4": 987.65,
		}
		assert.EqualValues(t, want, got)
	})
}

func TestEncodeAny(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   any
		want string
	}{
		{name: "nil", want: ""},
		{name: "string", in: "what is it?", want: "what is it?"},
		{name: "int", in: int64(123456789), want: "123456789"},
		{name: "float", in: 45.678, want: "45.678"},
		{name: "bool", in: true, want: "true"},
		{name: "encoder", in: &testEncoder{}, want: "hello world"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := encodeAny(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestEncodeBool(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   bool
		want string
	}{
		{name: "false", want: "false"},
		{name: "true", in: true, want: "true"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := encodeBool(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestEncodeTime(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		def  *schema.Field
		in   time.Time
		want string
	}{
		{
			name: "zero",
			def:  &schema.Field{},
			want: "",
		},
		{
			name: "BRP date",
			def:  &schema.Field{Format: "brp-date"},
			in:   time.Date(2025, 11, 12, 13, 14, 15, 0, time.UTC),
			want: "20251112",
		},
		{
			name: "formatted",
			def:  &schema.Field{Format: "02 01 2006"},
			in:   time.Date(2025, 11, 12, 13, 14, 15, 0, time.UTC),
			want: "12 11 2025",
		},
		{
			name: "no date",
			def:  &schema.Field{},
			in:   time.Date(0, 1, 1, 13, 14, 15, 0, time.UTC),
			want: "13:14:15",
		},
		{
			name: "no time",
			def:  &schema.Field{},
			in:   time.Date(2025, 12, 11, 0, 0, 0, 0, time.UTC),
			want: "2025-12-11",
		},
		{
			name: "no microseconds",
			def:  &schema.Field{},
			in:   time.Date(2025, 11, 12, 13, 14, 15, 0, time.UTC),
			want: "2025-11-12 13:14:15",
		},
		{
			name: "all",
			def:  &schema.Field{},
			in:   time.Date(2025, 11, 12, 13, 14, 15, 123456000, time.UTC),
			want: "2025-11-12 13:14:15.123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := encodeTime(tc.def, tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}

type testEncoder struct{}

func (e *testEncoder) Encode() any { return "hello world" }
