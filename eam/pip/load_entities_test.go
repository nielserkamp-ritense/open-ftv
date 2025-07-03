package pip

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestLoadEntityMap(t *testing.T) {
	t.Parallel()

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
						models.NewOriginalAttribute("int", 999, 999, ""),
						models.NewOriginalAttribute("hello", "world", "world", ""),
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
						models.NewOriginalAttribute("int", 999, 999, ""),
						models.NewOriginalAttribute("hello", "world", "world", ""),
					), "first", "second"),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			p := New(nil, logger)
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
	t.Parallel()

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
				models.NewEntity("user", "alice", models.NewAttributeSet(models.NewOriginalAttribute("security", "medium", "medium", ""))),
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
				models.NewEntity("user", "bob", models.NewAttributeSet(models.NewOriginalAttribute("security", "high", "high", ""))),
				models.NewEntity("user", "alice", models.NewAttributeSet(models.NewOriginalAttribute("security", "medium", "medium", ""))),
				models.NewEntity("user", "charlie", models.NewAttributeSet(models.NewOriginalAttribute("security", "low", "low", ""))),
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
				models.NewEntity("user", "bob", models.NewAttributeSet(models.NewOriginalAttribute("security", "high", "high", ""))),
				models.NewEntity("user", "alice", models.NewAttributeSet(models.NewOriginalAttribute("security", "medium", "medium", ""))),
				models.NewEntity("user", "charlie", models.NewAttributeSet(models.NewOriginalAttribute("security", "low", "low", ""))),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			p1 := New(nil, logger)
			require.NotNil(t, p1)

			h.Clear()

			p2, ok := p1.(*pip)
			require.True(t, ok)
			require.NotNil(t, p2)

			p2.loadEntitiesAny(tc.in)
			require.Zero(t, h.Count())

			tc.want.IterateEntities(func(e1 models.Entity) {
				e2 := p1.GetEntity(e1.UID())
				assert.True(t, models.EntityEqual(e1, e2))
			})

			p1.IterateEntities(func(e1 models.Entity) {
				e2 := tc.want.GetEntity(e1.UID())
				assert.True(t, models.EntityEqual(e1, e2))
			})
		})
	}
}

func TestLoadEntities(t *testing.T) {
	t.Parallel()

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
			path:    "../../testdata/unittest/pip/not_yaml.yaml",
			wantLog: 1,
			want:    models.NewEntitySet(),
		},
		{
			name: "1 entity - yaml",
			path: "../../testdata/unittest/pip/entities/entity.yaml",
			want: models.NewEntitySet(
				models.NewEntity("app", "app1", models.NewAttributeSet(
					models.NewOriginalAttribute("code", "app1", "app1", ""),
					models.NewOriginalAttribute("name", "App-1", "App-1", ""),
				)),
			),
		},
		{
			name: "1 entity - toml",
			path: "../../testdata/unittest/pip2/misc/entity.toml",
			want: models.NewEntitySet(
				models.NewEntity("app", "app1", models.NewAttributeSet(
					models.NewOriginalAttribute("code", "app1", "app1", ""),
					models.NewOriginalAttribute("name", "App-1", "App-1", ""),
				)),
			),
		},
		{
			name: "1 entity - json",
			path: "../../testdata/unittest/pip2/misc/entity.json",
			want: models.NewEntitySet(
				models.NewEntity("app", "app1", models.NewAttributeSet(
					models.NewOriginalAttribute("code", "app1", "app1", ""),
					models.NewOriginalAttribute("name", "App-1", "App-1", ""),
				)),
			),
		},
		{
			name: "1 entity - funny extension",
			path: "../../testdata/unittest/pip2/misc/entity.yaml-text",
			want: models.NewEntitySet(
				models.NewEntity("app", "app1", models.NewAttributeSet(
					models.NewOriginalAttribute("code", "app1", "app1", ""),
					models.NewOriginalAttribute("name", "App-1", "App-1", ""),
				)),
			),
		},
		{
			name: "1 entity - turtle",
			path: "../../testdata/unittest/pip2/misc/entity.ttl",
			want: models.NewEntitySet(
				models.NewEntity("app", "app1", models.NewAttributeSet(
					models.NewOriginalAttribute("code", "app1", "app1", ""),
					models.NewOriginalAttribute("name", "App-1", "App-1", ""),
				)),
			),
		},
		{
			name: "few entities - yaml",
			path: "../../testdata/unittest/pip/entities/entities.yaml",
			want: models.NewEntitySet(
				models.NewEntity("app", "app1", models.NewAttributeSet(
					models.NewOriginalAttribute("code", "app1", "app1", ""),
					models.NewOriginalAttribute("name", "App-1", "App-1", ""),
				)),
				models.NewEntity("app", "app2", models.NewAttributeSet(
					models.NewOriginalAttribute("code", "app2", "app2", ""),
					models.NewOriginalAttribute("name", "App-2", "App-2", ""),
				)),
				models.NewEntity("app", "app3", models.NewAttributeSet(
					models.NewOriginalAttribute("code", "app3", "app3", ""),
					models.NewOriginalAttribute("name", "App-3", "App-3", ""),
				), "app::app1", "app::app2",
				),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p := New(nil, logger).(*pip)

			h.Clear()

			p.loadEntities(tc.path)
			assert.Equal(t, tc.wantLog, h.Count())

			if tc.wantLog == 0 {
				tc.want.IterateEntities(func(e1 models.Entity) {
					e2 := p.GetEntity(e1.UID())
					require.NotNil(t, e2)
					assert.EqualValues(t, e1, e2)
				})

				p.IterateEntities(func(e1 models.Entity) {
					e2 := tc.want.GetEntity(e1.UID())
					require.NotNil(t, e2)
					assert.EqualValues(t, e1, e2)
				})
			}
		})
	}
}
