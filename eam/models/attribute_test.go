package models

import (
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/xsd"
)

func TestNewAttribute(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		key   string
		value any
	}{
		{name: "nil", key: "nil"},
		{name: "string", key: "s1", value: "v1"},
		{name: "int", key: "i1", value: 123},
		{name: "bool", key: "b1", value: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewAttribute(tc.key, tc.value)
			assert.Equal(t, tc.key, got.Key())
			assert.Equal(t, tc.value, got.Value())
			assert.Equal(t, tc.value, got.Original())
			assert.Empty(t, got.Type())
		})
	}
}

func TestNewAttributeWithType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		key       string
		value     any
		original  any
		t         string
		wantValue any
		wantJSON  string
	}{
		{name: "nil", key: "nil", wantJSON: `{"key":"nil","value":null}`},
		{name: "string", key: "s1", value: "v1", t: "string", wantValue: "v1", wantJSON: `{"key":"s1","value":"v1","type":"string"}`},
		{name: "int", key: "i1", value: 123, original: "123", t: "xsd:long", wantValue: int64(123), wantJSON: `{"key":"i1","value":123,"original":"123","type":"xsd:long"}`},
		{name: "bool", key: "b1", value: true, t: "xsd:boolean", wantValue: true, wantJSON: `{"key":"b1","value":true,"type":"xsd:boolean"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewAttributeWithType(tc.key, tc.value, tc.t)
			assert.Equal(t, tc.key, got.Key())
			assert.Equal(t, tc.wantValue, got.Value())
			assert.Equal(t, tc.value, got.Original())
			assert.Equal(t, tc.t, got.Type())

			if tc.original != nil {
				got.original = tc.original
			}

			b, err := json.Marshal(got)
			require.NoError(t, err)
			assert.Equal(t, tc.wantJSON, string(b))
		})
	}
}

func TestNewOriginalAttribute(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		key       string
		value     any
		original  any
		t         string
		wantValue any
		wantYAML  string
	}{
		{name: "nil", key: "nil", wantYAML: "key: nil\nvalue: null\n"},
		{name: "string", key: "s1", value: "v1", t: "string", wantValue: "v1", wantYAML: "key: s1\nvalue: v1\ntype: string\n"},
		{name: "int", key: "i1", value: 123, original: "123", t: "xsd:long", wantValue: int64(123), wantYAML: "key: i1\nvalue: 123\noriginal: \"123\"\ntype: xsd:long\n"},
		{name: "bool", key: "b1", value: true, original: "1", t: "xsd:boolean", wantValue: true, wantYAML: "key: b1\nvalue: true\noriginal: \"1\"\ntype: xsd:boolean\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewOriginalAttribute(tc.key, tc.value, tc.original, tc.t)
			assert.Equal(t, tc.key, got.Key())
			assert.Equal(t, tc.wantValue, got.Value())
			assert.Equal(t, tc.original, got.Original())
			assert.Equal(t, tc.t, got.Type())

			b, err := yaml.Marshal(got)
			require.NoError(t, err)
			assert.Equal(t, tc.wantYAML, string(b))
		})
	}
}

func TestAttribute_WithTitle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		in    *Attribute
		title string
	}{
		{
			name:  "no title",
			in:    &Attribute{},
			title: "New title",
		},
		{
			name:  "existing title",
			in:    &Attribute{title: "Old title"},
			title: "New title 2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.WithTitle(tc.title)
			require.NotNil(t, got)
			assert.Equal(t, tc.title, got.Title())
		})
	}
}

func TestAttribute_WithDescription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *Attribute
		desc string
	}{
		{
			name: "no description",
			in:   &Attribute{},
			desc: "New description",
		},
		{
			name: "existing description",
			in:   &Attribute{description: "Old description"},
			desc: "New description 2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.WithDescription(tc.desc)
			require.NotNil(t, got)
			assert.Equal(t, tc.desc, got.Description())
		})
	}
}

func TestAttribute_WithTags(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		key   string
		value any
		tags  []string
	}{
		{name: "single", key: "x1", value: 1, tags: []string{"x"}},
		{name: "few", key: "x2", value: 2, tags: []string{"x", "y", "z"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewAttribute(tc.key, tc.value).WithTags(tc.tags...)
			require.NotNil(t, got)
			assert.EqualValues(t, tc.tags, got.Tags())

			for i := range tc.tags {
				assert.True(t, got.HasTag(tc.tags[i]))
			}

			assert.False(t, got.HasTag("qqq"))
		})
	}
}

func TestAttribute_WithAudit(t *testing.T) {
	t.Parallel()

	now1 := time.Now().UTC()
	now2 := now1.Add(-13 * time.Hour)

	testCases := []struct {
		name  string
		key   string
		tp    string
		value any
		time1 time.Time
		user1 string
		time2 time.Time
		user2 string
	}{
		{name: "created", key: "a1", tp: "string", value: "hello bob", user1: "bob", time1: now1},
		{name: "updated", key: "a2", tp: "boolean", value: true, user2: "charlie", time2: now1},
		{name: "both", key: "a3", tp: "float", value: 3.14159265, user1: "alice", time1: now2, user2: "charlie", time2: now1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewAttributeWithType(tc.key, tc.value, tc.tp)
			require.NotNil(t, got)

			got.WithAudit(tc.time1, tc.user1, tc.time2, tc.user2)
			assert.EqualValues(t, tc.time1, got.Created())
			assert.EqualValues(t, tc.user1, got.CreatedBy())
			assert.EqualValues(t, tc.time2, got.Updated())
			assert.EqualValues(t, tc.user2, got.UpdatedBy())
		})
	}
}

func TestFromOAS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *attributes.Attribute
		want *Attribute
	}{
		{
			name: "empty",
			in:   &attributes.Attribute{},
			want: &Attribute{tags: make(map[string]struct{})},
		},
		{
			name: "partial",
			in:   &attributes.Attribute{Key: "a1", Value: 123, Type: "int"},
			want: &Attribute{key: "a1", value: int64(123), original: 123, tp: "int", tags: make(map[string]struct{})},
		},
		{
			name: "full",
			in: &attributes.Attribute{Key: "a2", Value: true, Type: "bool", Metadata: attributes.Metadata{
				Description: "this is attribute a2",
				Tags:        []string{"x", "y", "z"},
				Title:       "attribute a2",
			}},
			want: &Attribute{
				key:         "a2",
				title:       "attribute a2",
				description: "this is attribute a2",
				value:       true,
				original:    true,
				tp:          "bool",
				tags:        map[string]struct{}{"x": {}, "y": {}, "z": {}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewAttributeFromOAS(tc.in)
			assert.True(t, tc.want.Equals(got))
		})
	}
}

func TestAttribute_ToBundle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *Attribute
		want *attributes.Attribute
	}{
		{
			name: "empty",
			in:   &Attribute{},
			want: &attributes.Attribute{},
		},
		{
			name: "simple",
			in:   NewAttribute("a1", "hello jupiter"),
			want: &attributes.Attribute{Key: "a1", Value: "hello jupiter", Type: "string"},
		},
		{
			name: "full",
			in: NewOriginalAttribute("a2", -12.34, "-12.34", "xsd:double").
				WithTitle("attribute a2").
				WithDescription("this is attribute a2").WithTags("x", "y", "z"),
			want: &attributes.Attribute{Key: "a2", Value: -12.34, Type: "xsd:double"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.ToBundle()
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestNewAttributeFromOAS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *attributes.Attribute
		want *Attribute
	}{
		{name: "no type - string", in: &attributes.Attribute{Key: "key", Value: "value"}, want: NewAttribute("key", "value")},
		{name: "no type - integer", in: &attributes.Attribute{Key: "key", Value: 123}, want: NewAttribute("key", 123)},
		{name: "no type - bool", in: &attributes.Attribute{Key: "key", Value: true}, want: NewAttribute("key", true)},
		{name: "no type - double", in: &attributes.Attribute{Key: "key", Value: 22.33}, want: NewAttribute("key", 22.33)},
		{name: "no type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("999")}, want: NewAttribute("key", json.Number("999"))},
		{name: "string type - string", in: &attributes.Attribute{Key: "key", Value: "value", Type: "string"}, want: NewAttributeWithType("key", "value", "string")},
		{name: "string type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "string"}, want: NewAttributeWithType("key", "987", "string")},
		{name: "string type - int64", in: &attributes.Attribute{Key: "key", Value: int64(987), Type: "string"}, want: NewAttributeWithType("key", "987", "string")},
		{name: "string type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "string"}, want: NewAttributeWithType("key", "777", "string")},
		{name: "string type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "string"}, want: NewAttributeWithType("key", "987.432", "string")},
		{name: "string type - bool", in: &attributes.Attribute{Key: "key", Value: false, Type: "string"}, want: NewAttributeWithType("key", "false", "string")},
		{name: "string type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "string"}, want: NewAttributeWithType("key", "[1 2 3]", "string")},
		{name: "integer type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "integer"}, want: NewAttributeWithType("key", int64(0), "integer")},
		{name: "integer type - string good", in: &attributes.Attribute{Key: "key", Value: "123", Type: "integer"}, want: NewAttributeWithType("key", int64(123), "integer")},
		{name: "integer type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "integer"}, want: NewAttributeWithType("key", int64(987), "integer")},
		{name: "integer type - int64", in: &attributes.Attribute{Key: "key", Value: int64(987), Type: "integer"}, want: NewAttributeWithType("key", int64(987), "integer")},
		{name: "integer type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "integer"}, want: NewAttributeWithType("key", int64(777), "integer")},
		{name: "integer type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "integer"}, want: NewAttributeWithType("key", int64(987), "integer")},
		{name: "integer type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "integer"}, want: NewAttributeWithType("key", int64(1), "integer")},
		{name: "integer type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "integer"}, want: NewAttributeWithType("key", int64(0), "integer")},
		{name: "int type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "int"}, want: NewAttributeWithType("key", int64(0), "int")},
		{name: "int type - string good", in: &attributes.Attribute{Key: "key", Value: "123", Type: "int"}, want: NewAttributeWithType("key", int64(123), "int")},
		{name: "int type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "int"}, want: NewAttributeWithType("key", int64(987), "int")},
		{name: "int type - int64", in: &attributes.Attribute{Key: "key", Value: int64(987), Type: "int"}, want: NewAttributeWithType("key", int64(987), "int")},
		{name: "int type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "int"}, want: NewAttributeWithType("key", int64(777), "int")},
		{name: "int type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "int"}, want: NewAttributeWithType("key", int64(987), "int")},
		{name: "int type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "int"}, want: NewAttributeWithType("key", int64(1), "int")},
		{name: "int type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "int"}, want: NewAttributeWithType("key", int64(0), "int")},
		{name: "short type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "short"}, want: NewAttributeWithType("key", int64(0), "short")},
		{name: "short type - string good", in: &attributes.Attribute{Key: "key", Value: "123", Type: "short"}, want: NewAttributeWithType("key", int64(123), "short")},
		{name: "long type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "long"}, want: NewAttributeWithType("key", int64(987), "long")},
		{name: "long type - int64", in: &attributes.Attribute{Key: "key", Value: int64(987), Type: "long"}, want: NewAttributeWithType("key", int64(987), "long")},
		{name: "long type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "long"}, want: NewAttributeWithType("key", int64(777), "long")},
		{name: "byte type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "byte"}, want: NewAttributeWithType("key", int64(987), "byte")},
		{name: "byte type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "byte"}, want: NewAttributeWithType("key", int64(1), "byte")},
		{name: "byte type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "byte"}, want: NewAttributeWithType("key", int64(0), "byte")},
		{name: "boolean type - string bad", in: &attributes.Attribute{Key: "key", Value: "yeah", Type: "boolean"}, want: NewAttributeWithType("key", false, "boolean")},
		{name: "boolean type - string good", in: &attributes.Attribute{Key: "key", Value: "true", Type: "boolean"}, want: NewAttributeWithType("key", true, "boolean")},
		{name: "boolean type - int", in: &attributes.Attribute{Key: "key", Value: 1, Type: "boolean"}, want: NewAttributeWithType("key", true, "boolean")},
		{name: "boolean type - int64", in: &attributes.Attribute{Key: "key", Value: int64(1), Type: "boolean"}, want: NewAttributeWithType("key", true, "boolean")},
		{name: "boolean type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("0"), Type: "boolean"}, want: NewAttributeWithType("key", false, "boolean")},
		{name: "boolean type - double", in: &attributes.Attribute{Key: "key", Value: 1.1, Type: "boolean"}, want: NewAttributeWithType("key", true, "boolean")},
		{name: "boolean type - bool", in: &attributes.Attribute{Key: "key", Value: false, Type: "boolean"}, want: NewAttributeWithType("key", false, "boolean")},
		{name: "boolean type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "boolean"}, want: NewAttributeWithType("key", false, "boolean")},
		{name: "bool type - string bad", in: &attributes.Attribute{Key: "key", Value: "yeah", Type: "bool"}, want: NewAttributeWithType("key", false, "bool")},
		{name: "bool type - string good", in: &attributes.Attribute{Key: "key", Value: "true", Type: "bool"}, want: NewAttributeWithType("key", true, "bool")},
		{name: "bool type - int", in: &attributes.Attribute{Key: "key", Value: 1, Type: "bool"}, want: NewAttributeWithType("key", true, "bool")},
		{name: "bool type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("0"), Type: "bool"}, want: NewAttributeWithType("key", false, "bool")},
		{name: "bool type - double", in: &attributes.Attribute{Key: "key", Value: 1.1, Type: "bool"}, want: NewAttributeWithType("key", true, "bool")},
		{name: "bool type - bool", in: &attributes.Attribute{Key: "key", Value: false, Type: "bool"}, want: NewAttributeWithType("key", false, "bool")},
		{name: "bool type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "bool"}, want: NewAttributeWithType("key", false, "bool")},
		{name: "double type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "double"}, want: NewAttributeWithType("key", float64(0), "double")},
		{name: "double type - string good", in: &attributes.Attribute{Key: "key", Value: "123.45678", Type: "double"}, want: NewAttributeWithType("key", 123.45678, "double")},
		{name: "double type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "double"}, want: NewAttributeWithType("key", float64(987), "double")},
		{name: "double type - int64", in: &attributes.Attribute{Key: "key", Value: int64(987), Type: "double"}, want: NewAttributeWithType("key", float64(987), "double")},
		{name: "double type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "double"}, want: NewAttributeWithType("key", float64(777), "double")},
		{name: "double type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "double"}, want: NewAttributeWithType("key", 987.432, "double")},
		{name: "double type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "double"}, want: NewAttributeWithType("key", float64(1), "double")},
		{name: "double type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "double"}, want: NewAttributeWithType("key", float64(0), "double")},
		{name: "float type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "float"}, want: NewAttributeWithType("key", float64(0), "float")},
		{name: "float type - string good", in: &attributes.Attribute{Key: "key", Value: "123.45678", Type: "float"}, want: NewAttributeWithType("key", 123.45678, "float")},
		{name: "float type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "float"}, want: NewAttributeWithType("key", float64(987), "float")},
		{name: "float type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "float"}, want: NewAttributeWithType("key", float64(777), "float")},
		{name: "float type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "float"}, want: NewAttributeWithType("key", 987.432, "float")},
		{name: "float type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "float"}, want: NewAttributeWithType("key", float64(1), "float")},
		{name: "float type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "float"}, want: NewAttributeWithType("key", float64(0), "float")},
		{name: "date type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "date"}, want: NewAttributeWithType("key", time.Time{}, "date")},
		{name: "date type - string good", in: &attributes.Attribute{Key: "key", Value: "2024-07-21", Type: "date"}, want: NewAttributeWithType("key", time.Date(2024, 07, 21, 0, 0, 0, 0, time.UTC), "date")},
		{name: "date type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "date"}, want: NewAttributeWithType("key", time.Time{}, "date")},
		{name: "date type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "date"}, want: NewAttributeWithType("key", time.Time{}, "date")},
		{name: "date type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "date"}, want: NewAttributeWithType("key", time.Time{}, "date")},
		{name: "date type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "date"}, want: NewAttributeWithType("key", time.Time{}, "date")},
		{name: "date type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "date"}, want: NewAttributeWithType("key", time.Time{}, "date")},
		{name: "time type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "time"}, want: NewAttributeWithType("key", time.Time{}, "time")},
		{name: "time type - string good", in: &attributes.Attribute{Key: "key", Value: "17:31:58", Type: "time"}, want: NewAttributeWithType("key", time.Date(0, 1, 1, 17, 31, 58, 0, time.UTC), "time")},
		{name: "time type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "time"}, want: NewAttributeWithType("key", time.Time{}, "time")},
		{name: "time type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "time"}, want: NewAttributeWithType("key", time.Time{}, "time")},
		{name: "time type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "time"}, want: NewAttributeWithType("key", time.Time{}, "time")},
		{name: "time type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "time"}, want: NewAttributeWithType("key", time.Time{}, "time")},
		{name: "time type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "time"}, want: NewAttributeWithType("key", time.Time{}, "time")},
		{name: "dateTime type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "datetime"}, want: NewAttributeWithType("key", time.Time{}, "datetime")},
		{name: "dateTime type - string good", in: &attributes.Attribute{Key: "key", Value: "2024-07-21T07:59:11Z", Type: "datetime"}, want: NewAttributeWithType("key", time.Date(2024, 07, 21, 7, 59, 11, 0, time.UTC), "datetime")},
		{name: "dateTime type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "datetime"}, want: NewAttributeWithType("key", time.Time{}, "datetime")},
		{name: "dateTime type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "datetime"}, want: NewAttributeWithType("key", time.Time{}, "datetime")},
		{name: "dateTime type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "datetime"}, want: NewAttributeWithType("key", time.Time{}, "datetime")},
		{name: "dateTime type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "datetime"}, want: NewAttributeWithType("key", time.Time{}, "datetime")},
		{name: "dateTime type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "datetime"}, want: NewAttributeWithType("key", time.Time{}, "datetime")},
		{name: "timestamp type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "timestamp"}, want: NewAttributeWithType("key", time.Time{}, "timestamp")},
		{name: "timestamp type - string good", in: &attributes.Attribute{Key: "key", Value: "2024-07-21T07:59:11Z", Type: "timestamp"}, want: NewAttributeWithType("key", time.Date(2024, 07, 21, 7, 59, 11, 0, time.UTC), "timestamp")},
		{name: "timestamp type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "timestamp"}, want: NewAttributeWithType("key", time.Time{}, "timestamp")},
		{name: "timestamp type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "timestamp"}, want: NewAttributeWithType("key", time.Time{}, "timestamp")},
		{name: "timestamp type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "timestamp"}, want: NewAttributeWithType("key", time.Time{}, "timestamp")},
		{name: "timestamp type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "timestamp"}, want: NewAttributeWithType("key", time.Time{}, "timestamp")},
		{name: "timestamp type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "timestamp"}, want: NewAttributeWithType("key", time.Time{}, "timestamp")},
		{name: "xsd:any - string", in: &attributes.Attribute{Key: "key", Value: "value", Type: xsd.PrefixAny}, want: NewAttributeWithType("key", "value", "xsd:anyType")},
		{name: "xsd:any - int", in: &attributes.Attribute{Key: "key", Value: 654321, Type: xsd.PrefixAny}, want: NewAttributeWithType("key", 654321, "xsd:anyType")},
		{name: "xsd:any - double", in: &attributes.Attribute{Key: "key", Value: 555.888, Type: xsd.PrefixAny}, want: NewAttributeWithType("key", 555.888, "xsd:anyType")},
		{name: "xsd:any - json.number", in: &attributes.Attribute{Key: "key", Value: json.Number("765"), Type: xsd.PrefixAny}, want: NewAttributeWithType("key", json.Number("765"), "xsd:anyType")},
		{name: "xsd:anyURI - string", in: &attributes.Attribute{Key: "key", Value: "https://localhost:443", Type: xsd.PrefixAnyURI}, want: NewAttributeWithType("key", "https://localhost:443", "xsd:anyURI")},
		{name: "xsd:bool type - string bad", in: &attributes.Attribute{Key: "key", Value: "yeah", Type: xsd.PrefixBoolean}, want: NewAttributeWithType("key", false, "xsd:boolean")},
		{name: "xsd:bool type - string good", in: &attributes.Attribute{Key: "key", Value: "true", Type: xsd.PrefixBoolean}, want: NewAttributeWithType("key", true, "xsd:boolean")},
		{name: "xsd:string - string", in: &attributes.Attribute{Key: "key", Value: "value", Type: xsd.PrefixString}, want: NewAttributeWithType("key", "value", "xsd:string")},
		{name: "xsd:duration type - string bad", in: &attributes.Attribute{Key: "key", Value: "xyz", Type: xsd.PrefixDuration}, want: NewAttributeWithType("key", time.Duration(0), "xsd:duration")},
		{name: "xsd:duration type - string good (1)", in: &attributes.Attribute{Key: "key", Value: "260s", Type: xsd.PrefixDuration}, want: NewAttributeWithType("key", 260*time.Second, "xsd:duration")},
		{name: "xsd:duration type - string good (2)", in: &attributes.Attribute{Key: "key", Value: "PT4M20S", Type: xsd.PrefixDuration}, want: NewAttributeWithType("key", 260*time.Second, "xsd:duration")},
		{name: "xsd:duration type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: xsd.PrefixDuration}, want: NewAttributeWithType("key", time.Duration(0), "xsd:duration")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewAttributeFromOAS(tc.in)
			require.NotNil(t, got)
			assert.Equal(t, tc.want.Key(), got.Key())
			assert.Equal(t, tc.want.Value(), got.Value())
			assert.Equal(t, tc.in.Value, got.Original())
			assert.Equal(t, tc.want.Type(), got.Type())
		})
	}
}

func TestAttribute_ToOAS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *Attribute
		want *attributes.Attribute
	}{
		{
			name: "empty",
			in:   &Attribute{},
			want: &attributes.Attribute{Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "simple",
			in:   NewAttribute("a1", "hello jupiter"),
			want: &attributes.Attribute{Key: "a1", Value: "hello jupiter", Type: "string", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "full",
			in: NewOriginalAttribute("a2", -12.34, "-12.34", "xsd:double").
				WithTitle("attribute a2").
				WithDescription("this is attribute a2").WithTags("x", "y", "z"),
			want: &attributes.Attribute{
				Key:   "a2",
				Value: -12.34,
				Type:  "xsd:double",
				Metadata: attributes.Metadata{
					Title:       "attribute a2",
					Description: "this is attribute a2",
					Tags:        []string{"x", "y", "z"},
				},
			},
		},
		{
			name: "string",
			in:   NewAttribute("key", "value"),
			want: &attributes.Attribute{Key: "key", Value: "value", Type: "string", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "int64",
			in:   NewAttribute("key", int64(9876543)),
			want: &attributes.Attribute{Key: "key", Value: int64(9876543), Type: "long", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "double",
			in:   NewAttribute("key", 123.5678),
			want: &attributes.Attribute{Key: "key", Value: 123.5678, Type: "double", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "bool",
			in:   NewAttribute("key", true),
			want: &attributes.Attribute{Key: "key", Value: true, Type: "bool", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "slice",
			in:   NewAttribute("key", []any{5, 6, 7, 8}),
			want: &attributes.Attribute{Key: "key", Value: []any{5, 6, 7, 8}, Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "map",
			in:   NewAttribute("key", map[string]any{"hello": "world", "int": 123}),
			want: &attributes.Attribute{Key: "key", Value: map[string]any{"hello": "world", "int": 123}, Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "time.Time",
			in:   NewAttribute("key", time.Date(2024, 5, 29, 12, 0, 0, 0, time.UTC)),
			want: &attributes.Attribute{Key: "key", Value: "2024-05-29T12:00:00Z", Type: "datetime", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "time.Duration",
			in:   NewAttribute("key", time.Duration(123456780000)),
			want: &attributes.Attribute{Key: "key", Value: "2m3.45678s", Type: "duration", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "unsupported",
			in:   NewAttribute("key", []float32{1, 2, 3}),
			want: &attributes.Attribute{Key: "key", Value: "[1 2 3]", Type: "string", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "with string type",
			in:   NewAttributeWithType("key", "true", "string"),
			want: &attributes.Attribute{Key: "key", Value: "true", Type: "string", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "with xsd:boolean type",
			in:   NewAttributeWithType("key", true, "xsd:boolean"),
			want: &attributes.Attribute{Key: "key", Value: true, Type: "xsd:boolean", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "with xsd:gYear type",
			in:   NewAttributeWithType("key", 2025, "xsd:gYear"),
			want: &attributes.Attribute{Key: "key", Value: int64(2025), Type: "xsd:gYear", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "with date type",
			in:   NewAttributeWithType("key", time.Date(2024, 5, 29, 12, 0, 0, 0, time.UTC), "date"),
			want: &attributes.Attribute{Key: "key", Value: "2024-05-29", Type: "date", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "with xsd:time type",
			in:   NewAttributeWithType("key", time.Date(2024, 5, 29, 12, 0, 0, 0, time.UTC), "xsd:time"),
			want: &attributes.Attribute{Key: "key", Value: "12:00:00", Type: "xsd:time", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "with xsd:dateTime type",
			in:   NewAttributeWithType("key", time.Date(2024, 5, 29, 12, 0, 0, 0, time.UTC), "xsd:dateTime"),
			want: &attributes.Attribute{Key: "key", Value: "2024-05-29T12:00:00Z", Type: "xsd:dateTime", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "with xsd:duration type (1)",
			in:   NewAttributeWithType("key", 260*time.Second, "xsd:duration"),
			want: &attributes.Attribute{Key: "key", Value: "PT4M20S", Type: "xsd:duration", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "with xsd:duration type (2)",
			in:   NewAttributeWithType("key", 300*time.Millisecond, "xsd:duration"),
			want: &attributes.Attribute{Key: "key", Value: "PT0.3S", Type: "xsd:duration", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "with xsd:duration type (3)",
			in:   NewAttributeWithType("key", time.Duration(1200), "xsd:duration"),
			want: &attributes.Attribute{Key: "key", Value: "PT0.0000012S", Type: "xsd:duration", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "with duration type",
			in:   NewAttributeWithType("key", time.Duration(123456780000), "duration"),
			want: &attributes.Attribute{Key: "key", Value: "2m3.45678s", Type: "duration", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "with original xsd:boolean",
			in:   NewOriginalAttribute("key", true, "1", "xsd:boolean"),
			want: &attributes.Attribute{Key: "key", Value: true, Type: "xsd:boolean", Metadata: attributes.Metadata{Tags: []string{}}},
		},
		{
			name: "with original xsd:duration",
			in:   NewOriginalAttribute("key", 260*time.Second, "PT4M20S", "xsd:duration"),
			want: &attributes.Attribute{Key: "key", Value: "PT4M20S", Type: "xsd:duration", Metadata: attributes.Metadata{Tags: []string{}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.ToOAS()
			require.NotNil(t, got)
			assert.EqualValues(t, tc.want, got)
		})
	}
}
