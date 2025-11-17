package models

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

func TestNewAttributeSet(t *testing.T) {
	t.Parallel()

	m1 := map[string]any{"hello": "world2", "int": 678}
	m2 := map[string]any{"bool": true}

	testCases := []struct {
		name string
		in   []any
		want map[string]any
	}{
		{
			name: "empty",
			want: make(map[string]any),
		},
		{
			name: "one set",
			in:   []any{NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 123))},
			want: map[string]any{"hello": "world", "int": 123},
		},
		{
			name: "map",
			in:   []any{m1},
			want: map[string]any{"hello": "world2", "int": 678},
		},
		{
			name: "*map",
			in:   []any{&m2},
			want: map[string]any{"bool": true},
		},
		{
			name: "mixed input",
			in: []any{
				NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 123)),
				12345,
				&m1,
				NewAttribute("world", "hello"),
				m2,
				NewAttributeSet(NewAttribute("hello", "world2"), NewAttribute("bool", true)),
				true,
				NewAttribute("int", 456),
				nil,
			},
			want: map[string]any{"hello": "world2", "int": 456, "bool": true, "world": "hello"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewAttributeSet(tc.in...)
			require.NotNil(t, got)

			got2 := MapFromAttributes(got)
			require.NotNil(t, got2)

			for k, v := range tc.want {
				v2 := got.GetAttributeValue(k)
				assert.Equal(t, v, v2)
			}

			got.IterateAttributes(func(attr *Attribute) {
				assert.Equal(t, tc.want[attr.Key()], attr.Value())
			})

			bad := got.GetAttribute("does.not.exist.dude")
			assert.Nil(t, bad)

			badValue := got.GetAttributeValue("does.not.exist.dude")
			assert.Nil(t, badValue)
		})
	}
}

func TestAttributes_AddAttribute(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		in    *AttributeSet
		key   string
		value any
		want  map[string]any
	}{
		{
			name:  "empty",
			in:    &AttributeSet{set: make(map[string]*Attribute)},
			key:   "hello",
			value: 12345,
			want:  map[string]any{"hello": 12345},
		},
		{
			name:  "new key",
			in:    NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 123), NewAttribute("bool", true)),
			key:   "float",
			value: 12345.6789,
			want:  map[string]any{"hello": "world", "int": 123, "bool": true, "float": 12345.6789},
		},
		{
			name:  "duplicate key",
			in:    NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 123), NewAttribute("bool", true)),
			key:   "int",
			value: 12345.6789,
			want:  map[string]any{"hello": "world", "bool": true, "int": 12345.6789},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := tc.in
			a.AddAttributeKV(tc.key, tc.value)

			got := MapFromAttributes(a)
			require.NotNil(t, got)
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestAttributes_AddAttributeWithType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		in        *AttributeSet
		key       string
		value     any
		tp        string
		wantValue any
		want      map[string]any
	}{
		{
			name:      "empty",
			in:        &AttributeSet{set: make(map[string]*Attribute)},
			key:       "hello",
			value:     12345,
			tp:        "xsd:integer",
			wantValue: int64(12345),
			want:      map[string]any{"hello": int64(12345)},
		},
		{
			name:      "new key",
			in:        NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 123), NewAttribute("bool", true)),
			key:       "float",
			value:     12345.6789,
			tp:        "xsd:double",
			wantValue: 12345.6789,
			want:      map[string]any{"hello": "world", "int": 123, "bool": true, "float": 12345.6789},
		},
		{
			name:      "duplicate key",
			in:        NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 123), NewAttribute("bool", true)),
			key:       "int",
			value:     12345.6789,
			tp:        "xsd:double",
			wantValue: 12345.6789,
			want:      map[string]any{"hello": "world", "bool": true, "int": 12345.6789},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := tc.in
			a.AddAttributeKVWithType(tc.key, tc.value, tc.tp)

			got := MapFromAttributes(a)
			require.NotNil(t, got)
			assert.EqualValues(t, tc.want, got)

			attr := a.GetAttribute(tc.key)
			require.NotNil(t, attr)
			assert.Equal(t, tc.wantValue, attr.Value())
			assert.Equal(t, tc.tp, attr.Type())
		})
	}
}

func TestAttributes_AddOriginalAttribute(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		in        *AttributeSet
		key       string
		value     any
		orig      any
		tp        string
		wantValue any
		want      map[string]any
	}{
		{
			name:      "empty",
			in:        &AttributeSet{set: make(map[string]*Attribute)},
			key:       "hello",
			value:     12345,
			orig:      int32(12345),
			tp:        "xsd:integer",
			wantValue: int64(12345),
			want:      map[string]any{"hello": int64(12345)},
		},
		{
			name:  "blank key",
			in:    &AttributeSet{set: make(map[string]*Attribute)},
			key:   "",
			value: 12345,
			orig:  int32(12345),
			tp:    "xsd:integer",
			want:  map[string]any{},
		},
		{
			name: "new key",
			in: NewAttributeSet(
				NewAttribute("hello", "world"),
				NewAttribute("int", 123),
				NewAttribute("bool", true),
			),
			key:       "float",
			value:     12345.6789,
			orig:      "12345.6789",
			tp:        "xsd:double",
			wantValue: 12345.6789,
			want:      map[string]any{"hello": "world", "int": 123, "bool": true, "float": 12345.6789},
		},
		{
			name: "duplicate key",
			in: NewAttributeSet(
				NewAttribute("hello", "world"),
				NewAttribute("int", 123),
				NewAttribute("bool", true),
			),
			key:       "int",
			value:     12345.6789,
			orig:      "12345.6789",
			tp:        "xsd:double",
			wantValue: 12345.6789,
			want:      map[string]any{"hello": "world", "bool": true, "int": 12345.6789},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := tc.in
			a.AddOriginalAttribute(tc.key, tc.value, tc.orig, tc.tp)

			got := MapFromAttributes(a)
			require.NotNil(t, got)
			assert.EqualValues(t, tc.want, got)

			if tc.key != "" {
				attr := a.GetAttribute(tc.key)
				require.NotNil(t, attr)
				assert.Equal(t, tc.wantValue, attr.Value())
				assert.Equal(t, tc.orig, attr.Original())
				assert.Equal(t, tc.tp, attr.Type())
			}
		})
	}
}

func TestAttributes_RemoveAttribute(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *AttributeSet
		key  string
		want map[string]any
	}{
		{
			name: "empty",
			in:   &AttributeSet{set: make(map[string]*Attribute)},
			key:  "hello",
			want: map[string]any{},
		},
		{
			name: "miss",
			in:   NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 123), NewAttribute("bool", true)),
			key:  "float",
			want: map[string]any{"hello": "world", "int": 123, "bool": true},
		},
		{
			name: "hit",
			in:   NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 123), NewAttribute("bool", true)),
			key:  "int",
			want: map[string]any{"hello": "world", "bool": true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := tc.in
			a.RemoveAttribute(tc.key)

			got := MapFromAttributes(a)
			require.NotNil(t, got)
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestAttributes_MergeAttributes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		in    *AttributeSet
		merge []*AttributeSet
		want  map[string]any
	}{
		{
			name: "both empty",
			in:   &AttributeSet{set: make(map[string]*Attribute)},
			want: make(map[string]any),
		},
		{
			name: "add one set",
			in:   NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 123)),
			merge: []*AttributeSet{
				NewAttributeSet(NewAttribute("hello", "world2"), NewAttribute("bool", true)),
			},
			want: map[string]any{"hello": "world2", "int": 123, "bool": true},
		},
		{
			name: "add few sets",
			in:   NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 234)),
			merge: []*AttributeSet{
				NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 123)),
				NewAttributeSet(NewAttribute("int", 456), NewAttribute("world", "hello")),
				NewAttributeSet(NewAttribute("hello", "world2"), NewAttribute("bool", true)),
			},
			want: map[string]any{"hello": "world2", "int": 456, "bool": true, "world": "hello"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := tc.in
			a.MergeAttributes(tc.merge...)

			got := MapFromAttributes(a)
			require.NotNil(t, got)
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestAttributes_MarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		attr []*Attribute
		want string
	}{
		{
			name: "empty",
			want: `[]`,
		},
		{
			name: "single",
			attr: []*Attribute{NewAttribute("hello", "world")},
			want: `[{"key":"hello","value":"world","status":"???"}]`,
		},
		{
			name: "few",
			attr: []*Attribute{
				NewAttribute("int", 999),
				NewAttribute("hello", "world"),
				NewAttribute("double", 123.456),
			},
			want: `[{"key":"double","value":123.456,"status":"???"},{"key":"hello","value":"world","status":"???"},{"key":"int","value":999,"status":"???"}]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := NewAttributeSet(tc.attr)
			require.NotNil(t, s)

			got, err := json.Marshal(s)
			require.NoError(t, err)
			assert.Equal(t, tc.want, string(got))
		})
	}
}

func TestMapFromAttributes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *AttributeSet
		want map[string]any
	}{
		{
			name: "nil",
		},
		{
			name: "empty",
			in:   NewAttributeSet(),
			want: map[string]any{},
		},
		{
			name: "few attributes",
			in:   NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 234)),
			want: map[string]any{"hello": "world", "int": 234},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := MapFromAttributes(tc.in)
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestAttributesEqual(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		s1   *AttributeSet
		s2   *AttributeSet
		want bool
	}{
		{
			name: "both nil",
			want: true,
		},
		{
			name: "both empty",
			s1:   NewAttributeSet(),
			s2:   NewAttributeSet(),
			want: true,
		},
		{
			name: "s1 empty",
			s1:   NewAttributeSet(),
			s2:   NewAttributeSet(NewAttribute("hello", "world")),
		},
		{
			name: "s2 empty",
			s1:   NewAttributeSet(NewAttribute("hello", "world")),
			s2:   NewAttributeSet(),
		},
		{
			name: "equal",
			s1: NewAttributeSet(
				NewAttribute("hello", "world"),
				NewAttribute("int", 123456),
				NewAttribute("bool", true),
			),
			s2: NewAttributeSet(
				NewAttribute("int", 123456),
				NewAttribute("bool", true),
				NewAttribute("hello", "world"),
			),
			want: true,
		},
		{
			name: "almost equal",
			s1: NewAttributeSet(
				NewAttribute("hello", "world"),
				NewAttribute("bool", true),
				NewAttribute("float", 123456.66),
			),
			s2: NewAttributeSet(
				NewAttribute("float", 123456.67),
				NewAttribute("bool", true),
				NewAttribute("hello", "world"),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.s1.Equals(tc.s2)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAttributeSet_ToOAS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *AttributeSet
		want []attributes.Attribute
	}{
		{
			name: "empty",
			in:   NewAttributeSet(),
			want: []attributes.Attribute{},
		},
		{
			name: "single",
			in:   NewAttributeSet(NewAttribute("hello", "world")),
			want: []attributes.Attribute{{Key: "hello", Value: "world", Status: "???", Type: "string", Metadata: attributes.Metadata{Tags: []string{}}}},
		},
		{
			name: "few",
			in: NewAttributeSet(
				NewAttribute("hello", "world"),
				NewAttribute("int", 123456),
				NewAttribute("bool", true),
				NewAttribute("float", 1.345),
			),
			want: []attributes.Attribute{
				{Key: "bool", Value: true, Status: "???", Type: "bool", Metadata: attributes.Metadata{Tags: []string{}}},
				{Key: "float", Value: 1.345, Status: "???", Type: "double", Metadata: attributes.Metadata{Tags: []string{}}},
				{Key: "hello", Value: "world", Status: "???", Type: "string", Metadata: attributes.Metadata{Tags: []string{}}},
				{Key: "int", Value: 123456, Status: "???", Metadata: attributes.Metadata{Tags: []string{}}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.ToOAS()
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestAttributeSeFromOAS(t *testing.T) {
	t.Parallel()

	a1 := attributes.Attribute{Key: "a1", Value: 123, Type: "xsd:integer"}
	a2 := attributes.Attribute{Key: "a2", Value: true, Type: "xsd:bool"}
	a3 := attributes.Attribute{Key: "a3", Value: "hello venus", Type: "xsd:string"}
	a4 := attributes.Attribute{Key: "a4", Value: 3.141592, Type: "xsd:double"}
	a5 := attributes.Attribute{Key: "a5", Value: -9999, Type: "xsd:integer"}
	a6 := attributes.Attribute{Key: "a6", Value: 1, Type: "xsd:bool"}

	testCases := []struct {
		name string
		add  []any
		want *AttributeSet
	}{
		{
			name: "none",
			want: NewAttributeSet(),
		},
		{
			name: "one attribute",
			add:  []any{a1},
			want: NewAttributeSet(NewAttributeFromOAS(&a1)),
		},
		{
			name: "pointer to one attribute",
			add:  []any{&a2},
			want: NewAttributeSet(NewAttributeFromOAS(&a2)),
		},
		{
			name: "list of attributes",
			add:  []any{[]attributes.Attribute{a3, a4, a5}},
			want: NewAttributeSet(NewAttributeFromOAS(&a3), NewAttributeFromOAS(&a5), NewAttributeFromOAS(&a4)),
		},
		{
			name: "list of pointers to attributes",
			add:  []any{[]*attributes.Attribute{&a4, &a2, &a6}},
			want: NewAttributeSet(NewAttributeFromOAS(&a2), NewAttributeFromOAS(&a4), NewAttributeFromOAS(&a6)),
		},
		{
			name: "combinations",
			add:  []any{a1, []*attributes.Attribute{&a4, &a2}, &a5, []attributes.Attribute{a6, a3}},
			want: NewAttributeSet(NewAttributeFromOAS(&a1), NewAttributeFromOAS(&a2), NewAttributeFromOAS(&a3), NewAttributeFromOAS(&a4), NewAttributeFromOAS(&a5), NewAttributeFromOAS(&a6)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := AttributeSetFromOAS(tc.add...)
			require.NotNil(t, got)
			assert.True(t, tc.want.Equals(got))
		})
	}
}
