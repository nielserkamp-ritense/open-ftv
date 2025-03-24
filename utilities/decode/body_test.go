package decode

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			a, err := parseJSON([]byte(tc.body))
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, a)
			} else {
				require.NoError(t, err)
				require.NotNil(t, a)
			}
		})
	}
}

func TestParseBody(t *testing.T) {
	testCases := []struct {
		name    string
		body    string
		ct      string
		wantErr bool
		want    map[string]any
	}{
		{
			name:    "empty - no content-type",
			wantErr: true,
		},
		{
			name:    "empty - content-type",
			ct:      "text/html; charset=utf-8",
			wantErr: true,
		},
		{
			name:    "unsupported - no content-type",
			body:    "hello, world!",
			wantErr: true,
		},
		{
			name:    "unsupported - content-type",
			body:    "hello, world!",
			ct:      "text/plain; charset=ebcdic",
			wantErr: true,
		},
		{
			name:    "invalid json - content-type",
			body:    "hello, world!",
			ct:      "text/json; charset=utf-32",
			wantErr: true,
		},
		{
			name:    "invalid xml - content-type",
			body:    "hello, world!",
			ct:      "text/xml; charset=utf-16",
			wantErr: true,
		},
		{
			name: "good json - no content-type",
			body: `{"hello":"world", "int": 123, "bool": true}`,
			want: map[string]any{
				"hello": "world",
				"int":   float64(123),
				"bool":  true,
			},
		},
		{
			name: "good json - content-type",
			body: `{"hello":"world", "int": 456, "bool": false}`,
			ct:   "application/json",
			want: map[string]any{
				"hello": "world",
				"int":   float64(456),
				"bool":  false,
			},
		},
		{
			name: "good xml - no content-type",
			body: "<start>hello world</start>",
			want: map[string]any{
				"nodes": []map[string]any{
					{
						"start": map[string]any{
							"attributes": []map[string]any{},
							"cdata":      "hello world",
							"nodes":      []map[string]any{},
						},
					},
				},
			},
		},
		{
			name: "good xml - content-type",
			body: "<start>world goodbye</start>",
			ct:   "application/xml",
			want: map[string]any{
				"nodes": []map[string]any{
					{
						"start": map[string]any{
							"attributes": []map[string]any{},
							"cdata":      "world goodbye",
							"nodes":      []map[string]any{},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseBody([]byte(tc.body), tc.ct)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.EqualValues(t, tc.want, got)
			}
		})
	}
}
