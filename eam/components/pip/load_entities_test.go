package pip

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestLoadEntityMap(t *testing.T) {
	testCases := []struct {
		name string
		in   map[string]any
		want models.EntitySet
	}{
		{
			name: "empty",
			want: models.NewEntitySet(),
		},
		{
			name: "simple",
			in:   map[string]any{"type": "service", "id": "myService"},
			want: models.NewEntitySet(
				models.NewEntity("service", "myService", models.NewAttributeSet()),
			),
		},
		{
			name: "with attributes",
			in: map[string]any{
				"type":       "service",
				"id":         "myService",
				"attributes": map[string]any{"hello": "world", "int": 999},
			},
			want: models.NewEntitySet(
				models.NewEntity("service", "myService",
					models.NewAttributeSet(
						models.NewAttribute("int", 999),
						models.NewAttribute("hello", "world"),
					),
				),
			),
		},
		{
			name: "with parents",
			in: map[string]any{
				"type":    "service",
				"id":      "myService",
				"parents": []any{123, "first", "second", true},
			},
			want: models.NewEntitySet(
				models.NewEntity("service", "myService", models.NewAttributeSet(), "first", "second"),
			),
		},
		{
			name: "with all",
			in: map[string]any{
				"attributes": map[string]any{"hello": "world", "int": 999},
				"id":         "myService",
				"parents":    []any{"first", 9.9, "second", 123},
				"type":       "service",
			},
			want: models.NewEntitySet(
				models.NewEntity("service", "myService",
					models.NewAttributeSet(
						models.NewAttribute("int", 999),
						models.NewAttribute("hello", "world"),
					), "first", "second"),
			),
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

			h.Clear()

			p2, ok := p.(*pip)
			require.True(t, ok)
			require.NotNil(t, p2)

			p2.loadEntityMap(tc.in)
			require.Zero(t, h.Count())

			tc.want.IterateEntities(func(e1 models.Entity) {
				e2 := p.GetEntity(e1.UID())
				assert.EqualValues(t, e1, e2)
			})

			p.IterateEntities(func(e1 models.Entity) {
				e2 := tc.want.GetEntity(e1.UID())
				assert.EqualValues(t, e1, e2)
			})
		})
	}
}

func TestLoadEntitiesAny(t *testing.T) {
	testCases := []struct {
		name string
		in   any
		want models.EntitySet
	}{
		{
			name: "nil",
			want: models.NewEntitySet(),
		},
		{
			name: "empty",
			in:   map[string]any{},
			want: models.NewEntitySet(),
		},
		{
			name: "single",
			in:   map[string]any{"type": "user", "id": "alice", "attributes": map[string]any{"security": "medium"}},
			want: models.NewEntitySet(
				models.NewEntity("user", "alice", models.NewAttributeSet(models.NewAttribute("security", "medium"))),
			),
		},
		{
			name: "few (1)",
			in: []map[string]any{
				{"type": "user", "id": "alice", "attributes": map[string]any{"security": "medium"}},
				{"type": "user", "id": "charlie", "attributes": map[string]any{"security": "low"}},
				{"type": "user", "id": "bob", "attributes": map[string]any{"security": "high"}},
			},
			want: models.NewEntitySet(
				models.NewEntity("user", "bob", models.NewAttributeSet(models.NewAttribute("security", "high"))),
				models.NewEntity("user", "alice", models.NewAttributeSet(models.NewAttribute("security", "medium"))),
				models.NewEntity("user", "charlie", models.NewAttributeSet(models.NewAttribute("security", "low"))),
			),
		},
		{
			name: "few (2)",
			in: []any{
				map[string]any{"type": "user", "id": "alice", "attributes": map[string]any{"security": "medium"}},
				map[string]any{"type": "user", "id": "charlie", "attributes": map[string]any{"security": "low"}},
				map[string]any{"type": "user", "id": "bob", "attributes": map[string]any{"security": "high"}},
			},
			want: models.NewEntitySet(
				models.NewEntity("user", "bob", models.NewAttributeSet(models.NewAttribute("security", "high"))),
				models.NewEntity("user", "alice", models.NewAttributeSet(models.NewAttribute("security", "medium"))),
				models.NewEntity("user", "charlie", models.NewAttributeSet(models.NewAttribute("security", "low"))),
			),
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

			h.Clear()

			p2, ok := p.(*pip)
			require.True(t, ok)
			require.NotNil(t, p2)

			p2.loadEntitiesAny(tc.in)
			require.Zero(t, h.Count())

			tc.want.IterateEntities(func(e1 models.Entity) {
				e2 := p.GetEntity(e1.UID())
				assert.EqualValues(t, e1, e2)
			})

			p.IterateEntities(func(e1 models.Entity) {
				e2 := tc.want.GetEntity(e1.UID())
				assert.EqualValues(t, e1, e2)
			})
		})
	}
}

func TestLoadEntities(t *testing.T) {
	testCases := []struct {
		name    string
		path    string
		wantLog int
		want    models.EntitySet
	}{
		{
			name:    "empty",
			wantLog: 1,
		},
		{
			name:    "bad path",
			path:    "/this/is/another/bad/file",
			wantLog: 1,
			want:    models.NewEntitySet(),
		},
		{
			name:    "not yaml",
			path:    "../../../testdata/unittest/pip/not_yaml.yaml",
			wantLog: 1,
			want:    models.NewEntitySet(),
		},
		{
			name: "1 entity - yaml",
			path: "../../../testdata/unittest/pip/entities/entity.yaml",
			want: models.NewEntitySet(
				models.NewEntity("app", "app1", models.NewAttributeSet(
					models.NewAttribute("code", "app1"),
					models.NewAttribute("name", "App-1"),
				)),
			),
		},
		{
			name: "1 entity - toml",
			path: "../../../testdata/unittest/pip2/misc/entity.toml",
			want: models.NewEntitySet(
				models.NewEntity("app", "app1", models.NewAttributeSet(
					models.NewAttribute("code", "app1"),
					models.NewAttribute("name", "App-1"),
				)),
			),
		},
		{
			name: "1 entity - json",
			path: "../../../testdata/unittest/pip2/misc/entity.json",
			want: models.NewEntitySet(
				models.NewEntity("app", "app1", models.NewAttributeSet(
					models.NewAttribute("code", "app1"),
					models.NewAttribute("name", "App-1"),
				)),
			),
		},
		{
			name: "1 entity - funny extension",
			path: "../../../testdata/unittest/pip2/misc/entity.yaml-text",
			want: models.NewEntitySet(
				models.NewEntity("app", "app1", models.NewAttributeSet(
					models.NewAttribute("code", "app1"),
					models.NewAttribute("name", "App-1"),
				)),
			),
		},
		{
			name: "1 entity - turtle",
			path: "../../../testdata/unittest/pip2/misc/entity.ttl",
			want: models.NewEntitySet(
				models.NewEntity("app", "app1", models.NewAttributeSet(
					models.NewAttribute("code", "app1"),
					models.NewAttribute("name", "App-1"),
				)),
			),
		},
		{
			name: "few entities - yaml",
			path: "../../../testdata/unittest/pip/entities/entities.yaml",
			want: models.NewEntitySet(
				models.NewEntity("app", "app1", models.NewAttributeSet(
					models.NewAttribute("code", "app1"),
					models.NewAttribute("name", "App-1"),
				)),
				models.NewEntity("app", "app2", models.NewAttributeSet(
					models.NewAttribute("code", "app2"),
					models.NewAttribute("name", "App-2"),
				)),
				models.NewEntity("app", "app3", models.NewAttributeSet(
					models.NewAttribute("code", "app3"),
					models.NewAttribute("name", "App-3"),
				), "app::app1", "app::app2",
				),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)

			p := &pip{
				logger:        slog.New(h),
				entities:      models.NewEntitySet(),
				newAttributes: models.NewAttributeSet,
			}

			p.loadEntities(tc.path)
			assert.Equal(t, tc.wantLog, h.Count())

			if tc.wantLog == 0 {
				tc.want.IterateEntities(func(e1 models.Entity) {
					e2 := p.entities.GetEntity(e1.UID())
					require.NotNil(t, e2)
					assert.EqualValues(t, e1, e2)
				})

				p.entities.IterateEntities(func(e1 models.Entity) {
					e2 := tc.want.GetEntity(e1.UID())
					require.NotNil(t, e2)
					assert.EqualValues(t, e1, e2)
				})
			}
		})
	}
}
