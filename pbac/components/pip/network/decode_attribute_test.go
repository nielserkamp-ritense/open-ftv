package network

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestProcessAttribute(t *testing.T) {
	testCases := []struct {
		name      string
		key       string
		value     any
		tp        string
		obj       *AttributesMapping
		wantCount int
		wantKey   string
		wantValue models.Attribute
	}{
		{
			name:      "no key",
			value:     "oops",
			obj:       &AttributesMapping{Base: "first"},
			wantCount: 1,
		},
		{
			name:      "no code",
			key:       "oops",
			obj:       &AttributesMapping{Base: "first"},
			wantCount: 1,
		},
		{
			name:      "no type",
			key:       "second",
			value:     123,
			obj:       &AttributesMapping{Base: "first"},
			wantKey:   "second",
			wantValue: models.NewOriginalAttribute("second", 123, 123, ""),
		},
		{
			name:      "all filled",
			key:       "third",
			value:     "hello world",
			tp:        "xsd:string",
			obj:       &AttributesMapping{Base: "first"},
			wantKey:   "third",
			wantValue: models.NewOriginalAttribute("third", "hello world", "hello world", "xsd:string"),
		},
		{
			name:      "xsd conversion",
			key:       "fourth",
			value:     "123.456",
			tp:        "xsd:double",
			obj:       &AttributesMapping{Base: "first"},
			wantKey:   "fourth",
			wantValue: models.NewOriginalAttribute("fourth", 123.456, "123.456", "xsd:double"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			attr := models.NewAttributeSet()

			r := &runner{logger: logger}
			r.processAttribute(tc.obj.Base, tc.key, tc.value, tc.tp, attr)

			if tc.wantCount > 0 {
				assert.Equal(t, tc.wantCount, h.Count())
			} else {
				got := attr.GetAttribute(tc.wantKey)
				assert.Equal(t, tc.wantValue, got)
			}
		})
	}
}

func TestDecodeAttributeMap(t *testing.T) {
	m1 := map[string]any{"hello": "world", "int": 123, "bool": true, "type": "xsd:short"}

	testCases := []struct {
		name      string
		m         map[string]any
		base      string
		obj       *AttributeMapping
		wantCount int
		wantKey   string
		wantValue models.Attribute
	}{
		{
			name:      "no key code, no key value",
			m:         m1,
			base:      "first",
			obj:       &AttributeMapping{ValueField: "value"},
			wantCount: 1,
		},
		{
			name:      "invalid key code, no key value",
			m:         m1,
			base:      "first",
			obj:       &AttributeMapping{KeyField: "abc", ValueField: "value"},
			wantCount: 1,
		},
		{
			name:      "no key code, key value",
			m:         m1,
			base:      "first",
			obj:       &AttributeMapping{KeyValue: "hello", ValueField: "int", TypeValue: "xsd:integer"},
			wantKey:   "hello",
			wantValue: models.NewOriginalAttribute("hello", int64(123), 123, "xsd:integer"),
		},
		{
			name:      "invalid key code, key value",
			m:         m1,
			base:      "first",
			obj:       &AttributeMapping{KeyField: "abc", KeyValue: "good", ValueField: "bool"},
			wantKey:   "good",
			wantValue: models.NewOriginalAttribute("good", true, true, ""),
		},
		{
			name:      "valid key code, key value",
			m:         m1,
			base:      "first",
			obj:       &AttributeMapping{KeyField: "hello", KeyValue: "good", ValueField: "int", TypeValue: "xsd:float"},
			wantKey:   "world",
			wantValue: models.NewOriginalAttribute("world", 123.0, 123, "xsd:float"),
		},
		{
			name:      "empty map",
			m:         map[string]any{},
			base:      "first",
			obj:       &AttributeMapping{KeyField: "hello", ValueField: "bool"},
			wantCount: 1,
		},
		{
			name:      "no value code",
			m:         m1,
			base:      "first",
			obj:       &AttributeMapping{KeyField: "int"},
			wantCount: 1,
		},
		{
			name:      "invalid value code",
			m:         m1,
			base:      "first",
			obj:       &AttributeMapping{KeyField: "int", ValueField: "oops"},
			wantCount: 1,
		},
		{
			name:      "valid value code",
			m:         m1,
			base:      "first",
			obj:       &AttributeMapping{KeyField: "hello", ValueField: "int"},
			wantKey:   "world",
			wantValue: models.NewOriginalAttribute("world", 123, 123, ""),
		},
		{
			name:      "no type code, type value",
			m:         m1,
			base:      "first",
			obj:       &AttributeMapping{KeyField: "hello", ValueField: "int", TypeValue: "xsd:double"},
			wantKey:   "world",
			wantValue: models.NewOriginalAttribute("world", 123.0, 123, "xsd:double"),
		},
		{
			name:      "invalid type code, type value",
			m:         m1,
			base:      "first",
			obj:       &AttributeMapping{KeyField: "hello", ValueField: "int", TypeField: "oops", TypeValue: "xsd:double"},
			wantKey:   "world",
			wantValue: models.NewOriginalAttribute("world", 123.0, 123, "xsd:double"),
		},
		{
			name:      "valid type code, type value",
			m:         m1,
			base:      "first",
			obj:       &AttributeMapping{KeyField: "hello", ValueField: "int", TypeField: "type", TypeValue: "xsd:double"},
			wantKey:   "world",
			wantValue: models.NewOriginalAttribute("world", int64(123), 123, "xsd:short"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			attr := models.NewAttributeSet()

			r := &runner{logger: logger}
			r.decodeAttributeMap(tc.base, tc.m, tc.obj, attr)

			if tc.wantCount > 0 {
				assert.Equal(t, tc.wantCount, h.Count())
			} else {
				require.Zero(t, h.Count())

				got := attr.GetAttribute(tc.wantKey)
				assert.Equal(t, tc.wantValue, got)
			}
		})
	}
}

func TestDecodeAttributeData(t *testing.T) {
	m1 := map[string]interface{}{"hello": "world"}
	m2 := map[string]any{"hello": "world", "int": 123, "bool": true, "type": "xsd:short"}
	m3 := map[string]any{"hello": "mars", "int": 321, "bool": false, "type": "xsd:string"}
	m4 := map[string]any{"hello": "jupiter", "int": true, "bool": 123}

	s1 := []any{m4, m2, m3}

	testCases := []struct {
		name      string
		data      any
		base      string
		obj       *AttributeMapping
		wantCount int
		want      map[string]models.Attribute
	}{
		{
			name: "value as-is",
			data: m1,
			base: "first",
			obj:  &AttributeMapping{KeyValue: "key", ValueAsIs: true},
			want: map[string]models.Attribute{"key": models.NewOriginalAttribute("key", m1, m1, "")},
		},
		{
			name: "map (1)",
			data: m2,
			base: "first",
			obj:  &AttributeMapping{KeyField: "hello", ValueField: "int", TypeField: "type"},
			want: map[string]models.Attribute{"world": models.NewOriginalAttribute("world", int64(123), 123, "xsd:short")},
		},
		{
			name: "map (2)",
			data: m3,
			base: "first",
			obj:  &AttributeMapping{KeyField: "hello", ValueField: "bool", TypeField: "type"},
			want: map[string]models.Attribute{"mars": models.NewOriginalAttribute("mars", "false", false, "xsd:string")},
		},
		{
			name: "slice",
			data: s1,
			base: "first",
			obj:  &AttributeMapping{KeyField: "hello", ValueField: "int", TypeField: "type"},
			want: map[string]models.Attribute{
				"world":   models.NewOriginalAttribute("world", int64(123), 123, "xsd:short"),
				"mars":    models.NewOriginalAttribute("mars", "321", 321, "xsd:string"),
				"jupiter": models.NewOriginalAttribute("jupiter", true, true, ""),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			attr := models.NewAttributeSet()

			r := &runner{logger: logger}
			r.decodeAttributeData(tc.base, tc.data, tc.obj, attr)

			if tc.wantCount > 0 {
				assert.Equal(t, tc.wantCount, h.Count())
			} else {
				require.Zero(t, h.Count())

				for k := range tc.want {
					got := attr.GetAttribute(k)
					assert.Equal(t, tc.want[k], got)
				}
			}
		})
	}
}

func TestDecodeAttribute(t *testing.T) {
	m1 := map[string]interface{}{"hello": "world"}
	m2 := map[string]any{"hello": "world", "int": 123, "bool": true, "type": "xsd:short"}
	m3 := map[string]any{"hello": "mars", "int": 321, "bool": false, "type": "xsd:string"}
	m4 := map[string]any{"hello": "jupiter", "int": true, "bool": 123}

	s1 := []any{m4, m2, m3}

	mm1 := map[string]any{"first": m1, "second": m2}
	mm2 := map[string]any{"first": m2, "second": m1}
	mm3 := map[string]any{"first": m3, "second": m3}
	mm4 := map[string]any{"first": s1, "second": m4}

	testCases := []struct {
		name      string
		data      any
		base      string
		obj       *AttributeMapping
		wantCount int
		want      map[string]models.Attribute
	}{
		{
			name: "value as-is",
			data: mm1,
			base: "first",
			obj:  &AttributeMapping{KeyValue: "key", ValueAsIs: true},
			want: map[string]models.Attribute{"key": models.NewOriginalAttribute("key", m1, m1, "")},
		},
		{
			name: "not map and not slice",
			data: map[string]any{"first": 987654321},
			base: "first",
			obj:  &AttributeMapping{KeyValue: "key", TypeValue: "xsd:string"},
			want: map[string]models.Attribute{"key": models.NewOriginalAttribute("key", "987654321", 987654321, "xsd:string")},
		},
		{
			name: "map (1)",
			data: mm2,
			base: "first",
			obj:  &AttributeMapping{KeyField: "hello", ValueField: "int", TypeField: "type"},
			want: map[string]models.Attribute{"world": models.NewOriginalAttribute("world", int64(123), 123, "xsd:short")},
		},
		{
			name: "map (2)",
			data: mm3,
			base: "first",
			obj:  &AttributeMapping{KeyField: "hello", ValueField: "bool", TypeField: "type"},
			want: map[string]models.Attribute{"mars": models.NewOriginalAttribute("mars", "false", false, "xsd:string")},
		},
		{
			name: "slice",
			data: mm4,
			base: "first",
			obj:  &AttributeMapping{KeyField: "hello", ValueField: "int", TypeField: "type"},
			want: map[string]models.Attribute{
				"world":   models.NewOriginalAttribute("world", int64(123), 123, "xsd:short"),
				"mars":    models.NewOriginalAttribute("mars", "321", 321, "xsd:string"),
				"jupiter": models.NewOriginalAttribute("jupiter", true, true, ""),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			attr := models.NewAttributeSet()

			r := &runner{logger: logger, data: tc.data, manager: &manager{logger: logger, attributes: attr}}
			r.decodeAttribute(&AttributesMapping{
				Base: tc.base,
				Map:  []*AttributeMapping{tc.obj},
			})

			if tc.wantCount > 0 {
				assert.Equal(t, tc.wantCount, h.Count())
			} else {
				require.Zero(t, h.Count())

				for k := range tc.want {
					got := attr.GetAttribute(k)
					assert.Equal(t, tc.want[k], got)
				}
			}
		})
	}
}
