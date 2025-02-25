package pep

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestParseJSON(t *testing.T) {
	testCases := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{name: "empty"},
		{name: "bad json", body: `this is not [] json`, wantErr: true},
		{name: "array", body: `["hello", "world", 123, true, "well"]`, wantErr: true},
		{name: "object", body: `{"hello": "world", "int": 123, "bool": true}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := models.NewAttributeSet()
			require.NotNil(t, a)

			err := parseJSON([]byte(tc.body), a)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProcessBody(t *testing.T) {
	testCases := []struct {
		name    string
		body    []byte
		attr    models.AttributeSet
		wantLog int
		want    models.AttributeSet
	}{
		{
			name: "empty",
			attr: models.NewAttributeSet(),
		},
		{
			name:    "no content-type, bad data",
			attr:    models.NewAttributeSet(),
			body:    []byte("hello world"),
			wantLog: 1,
		},
		{
			name:    "no content-type, invalid xml",
			attr:    models.NewAttributeSet(),
			body:    []byte("<xml? wat is dit dan?"),
			wantLog: 1,
		},
		{
			name: "no content-type, xml data",
			attr: models.NewAttributeSet(),
			body: []byte("<start>hello world</start>"),
			want: models.NewAttributeSet(
				models.NewAttribute(
					"body", []map[string]any{
						{"start": map[string]any{
							"attributes": []map[string]any{},
							"cdata":      "hello world",
							"nodes":      []map[string]any{},
						}}},
				),
			),
		},
		{
			name:    "no content-type, invalid json",
			attr:    models.NewAttributeSet(),
			body:    []byte(`{"veld1": wat is dit dan?`),
			wantLog: 1,
		},
		{
			name: "no content-type, json data",
			attr: models.NewAttributeSet(),
			body: []byte(`{"hello": "world"}`),
			want: models.NewAttributeSet(models.NewAttribute("body", map[string]any{"hello": "world"})),
		},
		{
			name:    "content-type unsupported",
			attr:    models.NewAttributeSet(models.NewAttribute("content-type", "x-iets")),
			body:    []byte("[]"),
			wantLog: 1,
		},
		{
			name:    "content-type xml, invalid xml",
			attr:    models.NewAttributeSet(models.NewAttribute("content-type", "application/xml")),
			body:    []byte("<xml? wat is dit dan?"),
			wantLog: 1,
		},
		{
			name: "content-type xml, valid xml",
			attr: models.NewAttributeSet(models.NewAttribute("content-type", "application/xml")),
			body: []byte("<start>hello world</start>"),
			want: models.NewAttributeSet(
				models.NewAttribute("content-type", "application/xml"),
				models.NewAttribute(
					"body", []map[string]any{
						{"start": map[string]any{
							"attributes": []map[string]any{},
							"cdata":      "hello world",
							"nodes":      []map[string]any{},
						}}},
				),
			),
		},
		{
			name:    "content-type json, invalid json",
			attr:    models.NewAttributeSet(models.NewAttribute("content-type", "application/json")),
			body:    []byte("[nee,toch}"),
			wantLog: 1,
		},
		{
			name: "content-type json, valid json",
			attr: models.NewAttributeSet(models.NewAttribute("content-type", "text/json")),
			body: []byte(`{"hello": "world"}`),
			want: models.NewAttributeSet(
				models.NewAttribute("content-type", "text/json"),
				models.NewAttribute("body", map[string]any{"hello": "world"}),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := util.NewDummyHandler(slog.LevelDebug)

			p := New(nil, slog.New(h))
			require.NotNil(t, p)

			h.Clear()

			c := &collector{
				debug:  true,
				logger: slog.New(h),
				req:    &models.HTTPRequest{Body: tc.body},
				parc:   &models.PARC{Context: tc.attr},
			}
			c.decodeBody()

			assert.Equal(t, tc.wantLog, h.Count())

			if tc.want != nil {
				tc.want.IterateAttributes(func(attr models.Attribute) {
					v2 := c.parc.Context.GetAttributeValue(attr.Key())
					assert.EqualValues(t, attr.Value(), v2)
				})

				c.parc.Context.IterateAttributes(func(attr models.Attribute) {
					v2 := tc.want.GetAttributeValue(attr.Key())
					assert.EqualValues(t, attr.Value(), v2)
				})
			}
		})
	}
}
