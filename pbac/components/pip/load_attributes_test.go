package pip

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestLoadAttributeMap(t *testing.T) {
	testCases := []struct {
		name string
		in   map[string]any
	}{
		{name: "empty", in: map[string]any{}},
		{name: "simple", in: map[string]any{"key": "hello", "value": 123.456}},
		{name: "complex", in: map[string]any{"value": map[string]any{"int": 123, "bool": true}, "key": "m"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelInfo)

			p := New(Config{
				Logger:        slog.New(h),
				NewAttributes: models.NewAttributeSet,
				NewEntities:   models.NewEntitySet,
			})
			require.NotNil(t, p)

			p2, ok := p.(*pip)
			require.True(t, ok)
			require.NotNil(t, p2)

			p2.loadAttributeMap(tc.in)

			if k, ok2 := tc.in["key"].(string); ok2 {
				got := p.GetAttribute(k)
				require.NotNil(t, got)
				assert.EqualValues(t, tc.in["value"], got)
			}
		})
	}
}

func TestLoadAttributesAny(t *testing.T) {
	testCases := []struct {
		name string
		in   any
		want map[string]any
	}{
		{name: "nil"},
		{name: "empty", in: map[string]any{}},
		{
			name: "single",
			in:   map[string]any{"key": "hello", "value": 123.456},
			want: map[string]any{"hello": 123.456},
		},
		{
			name: "few (1)",
			in: []any{
				map[string]any{"key": "hello", "value": 123.456},
				map[string]any{"value": map[string]any{"int": 123, "bool": true}, "key": "m"},
			},
			want: map[string]any{
				"hello": 123.456,
				"m":     map[string]any{"int": 123, "bool": true},
			},
		},
		{
			name: "few (2)",
			in: []map[string]any{
				{"value": map[string]any{"int": 123, "bool": true}, "key": "m"},
				{"key": "hello", "value": 123.456},
			},
			want: map[string]any{
				"hello": 123.456,
				"m":     map[string]any{"bool": true, "int": 123},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelInfo)

			p := New(Config{
				Logger:        slog.New(h),
				NewAttributes: models.NewAttributeSet,
				NewEntities:   models.NewEntitySet,
			})
			require.NotNil(t, p)

			p2, ok := p.(*pip)
			require.True(t, ok)
			require.NotNil(t, p2)

			p2.loadAttributesAny(tc.in)

			for k := range tc.want {
				got := p.GetAttribute(k)
				require.NotNil(t, got)
				assert.EqualValues(t, tc.want[k], got)
			}
		})
	}
}

func TestLoadAttributes(t *testing.T) {
	testCases := []struct {
		name    string
		path    string
		wantLog int
		want    models.AttributeSet
	}{
		{
			name:    "empty",
			wantLog: 1,
		},
		{
			name:    "bad path",
			path:    "/this/is/another/bad/file",
			wantLog: 1,
			want:    models.NewAttributeSet(),
		},
		{
			name:    "not yaml",
			path:    "../../../testdata/unittest/pip/not_yaml.yaml",
			wantLog: 1,
			want:    models.NewAttributeSet(),
		},
		{
			name: "1 attribute - yaml",
			path: "../../../testdata/unittest/pip/attributes/attribute.yaml",
			want: models.NewAttributeSet(
				models.NewAttribute("maandag", 1),
			),
		},
		{
			name: "1 attribute - funny extension",
			path: "../../../testdata/unittest/pip2/misc/attribute.yaml-text",
			want: models.NewAttributeSet(
				models.NewAttribute("maandag", 1),
			),
		},
		{
			name: "1 attribute - toml",
			path: "../../../testdata/unittest/pip2/misc/attribute.toml",
			want: models.NewAttributeSet(
				models.NewAttribute("hello", "world"),
			),
		},
		{
			name: "1 attribute - json",
			path: "../../../testdata/unittest/pip2/misc/attribute.json",
			want: models.NewAttributeSet(
				models.NewAttribute("float", 987.123),
			),
		},
		{
			name: "few attributes - yaml",
			path: "../../../testdata/unittest/pip/attributes/attributes.yaml",
			want: models.NewAttributeSet(
				models.NewAttribute("maandag", 1),
				models.NewAttribute("dinsdag", 2),
				models.NewAttribute("woensdag", 3),
				models.NewAttribute("donderdag", 4),
				models.NewAttribute("vrijdag", 5),
			),
		},
		{
			name: "complex attribute - turtle",
			path: "../../../testdata/unittest/pip2/misc/attributes.ttl",
			want: models.NewAttributeSet(
				models.NewAttribute("werktijden", map[string]any{
					"dinsdag":   int64(2),
					"donderdag": int64(4),
					"maandag":   int64(1),
					"vrijdag":   int64(5),
					"woensdag":  int64(3),
				}),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)

			p := &pip{
				logger:     slog.New(h),
				attributes: models.NewAttributeSet(),
			}

			p.loadAttributes(tc.path)

			assert.Equal(t, tc.wantLog, h.Count())

			if tc.wantLog == 0 {
				tc.want.IterateAttributes(func(key string, v1 any) {
					v2 := p.attributes.GetAttribute(key)
					assert.EqualValues(t, v1, v2)
				})

				p.attributes.IterateAttributes(func(key string, v1 any) {
					v2 := tc.want.GetAttribute(key)
					assert.EqualValues(t, v1, v2)
				})
			}
		})
	}
}
