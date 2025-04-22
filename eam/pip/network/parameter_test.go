package network

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/xsd"
)

func TestParameter_AttributeValue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		value      any
		tp         string
		attrKey    string
		get        models.GetAttribute
		want       any
		wantType   string
		wantStatic bool
	}{
		{
			name:       "invalid attribute key, no value",
			attrKey:    "abc",
			get:        &nilGetter{},
			want:       invalidValue,
			wantType:   xsd.PrefixString,
			wantStatic: true,
		},
		{
			name:       "invalid attribute key, value",
			value:      int64(123),
			tp:         xsd.PrefixInt,
			attrKey:    "abc",
			get:        &nilGetter{},
			want:       int64(123),
			wantType:   xsd.PrefixInt,
			wantStatic: true,
		},
		{
			name:     "valid attribute key, no value",
			attrKey:  "abc",
			get:      &fixedGetter{attr: models.NewAttributeWithType("abc", "hello world", xsd.PrefixString)},
			want:     "hello world",
			wantType: xsd.PrefixString,
		},
		{
			name:     "valid attribute key, value",
			value:    int64(123),
			tp:       xsd.PrefixShort,
			attrKey:  "abc",
			get:      &fixedGetter{attr: models.NewAttributeWithType("abc", 123.567, xsd.PrefixDouble)},
			want:     123.567,
			wantType: xsd.PrefixDouble,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &Parameter{Name: "param", Value: tc.value, Type: tc.tp, Attribute: tc.attrKey}

			got, tp, static := p.AttributeValue(tc.get)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantType, tp)
			assert.Equal(t, tc.wantStatic, static)
		})
	}
}

func TestParameter_ValueAny(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		value      any
		tp         string
		attrKey    string
		get        models.GetAttribute
		want       any
		wantType   string
		wantStatic bool
	}{
		{
			name:       "empty attribute key, no value",
			want:       invalidValue,
			wantType:   xsd.PrefixString,
			wantStatic: true,
		},
		{
			name:       "empty attribute key, value",
			value:      true,
			tp:         xsd.PrefixBoolean,
			want:       true,
			wantType:   xsd.PrefixBoolean,
			wantStatic: true,
		},
		{
			name:       "invalid attribute key, no value",
			attrKey:    "abc",
			get:        &nilGetter{},
			want:       invalidValue,
			wantType:   xsd.PrefixString,
			wantStatic: true,
		},
		{
			name:       "invalid attribute key, value",
			value:      int64(123),
			tp:         xsd.PrefixLong,
			attrKey:    "abc",
			get:        &nilGetter{},
			want:       int64(123),
			wantType:   xsd.PrefixLong,
			wantStatic: true,
		},
		{
			name:     "valid attribute key, no value",
			attrKey:  "abc",
			get:      &fixedGetter{attr: models.NewAttributeWithType("abc", "hello world", xsd.PrefixString)},
			want:     "hello world",
			wantType: xsd.PrefixString,
		},
		{
			name:     "valid attribute key, value",
			value:    int64(123),
			tp:       xsd.PrefixLong,
			attrKey:  "abc",
			get:      &fixedGetter{attr: models.NewAttributeWithType("abc", 123.567, xsd.PrefixDouble)},
			want:     123.567,
			wantType: xsd.PrefixDouble,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &Parameter{Name: "param", Value: tc.value, Type: tc.tp, Attribute: tc.attrKey}

			got, tp, static := p.ValueAny(tc.get)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantType, tp)
			assert.Equal(t, tc.wantStatic, static)
		})
	}
}

func TestParameter_ValueString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		value      any
		t          string
		attrKey    string
		get        models.GetAttribute
		want       string
		wantStatic bool
	}{
		{
			name:       "empty attribute key, no value",
			want:       invalidValue,
			wantStatic: true,
		},
		{
			name:       "empty attribute key, value",
			value:      true,
			t:          "xsd:boolean",
			want:       "true",
			wantStatic: true,
		},
		{
			name:       "invalid attribute key, no value",
			attrKey:    "abc",
			get:        &nilGetter{},
			want:       invalidValue,
			wantStatic: true,
		},
		{
			name:       "invalid attribute key, value",
			value:      int64(123),
			t:          "xsd:long",
			attrKey:    "abc",
			get:        &nilGetter{},
			want:       "123",
			wantStatic: true,
		},
		{
			name:    "valid attribute key, no value",
			t:       "xsd:string",
			attrKey: "abc",
			get:     &fixedGetter{attr: models.NewAttributeWithType("abc", "hello world", xsd.PrefixString)},
			want:    "hello world",
		},
		{
			name:    "valid attribute key, value",
			value:   int64(123),
			t:       "xsd:double",
			attrKey: "abc",
			get:     &fixedGetter{attr: models.NewAttributeWithType("abc", 123.567, xsd.PrefixDouble)},
			want:    "123.567",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &Parameter{Name: "param", Value: tc.value, Type: tc.t, Attribute: tc.attrKey}

			got, static := p.ValueString(tc.get)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantStatic, static)
		})
	}
}

type fixedGetter struct {
	attr models.Attribute
}

func (f *fixedGetter) GetAttribute(_ string) models.Attribute { return f.attr }

type nilGetter struct{}

func (f *nilGetter) GetAttribute(_ string) models.Attribute { return nil }
