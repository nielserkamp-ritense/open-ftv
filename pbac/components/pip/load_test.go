package pip

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

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
			name: "1 entity",
			path: "../../../testdata/unittest/pip/entities/entity.yaml",
			want: models.NewEntitySet(
				models.NewEntity("app", "app1", models.NewAttributeSet(
					models.NewAttribute("code", "app1"),
					models.NewAttribute("name", "App-1"),
				)),
			),
		},
		{
			name: "few entities",
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
			h := util.NewDummyHandler(slog.LevelDebug)

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
			name: "1 attribute",
			path: "../../../testdata/unittest/pip/attributes/attribute.yaml",
			want: models.NewAttributeSet(
				models.NewAttribute("maandag", 1),
			),
		},
		{
			name: "few attributes",
			path: "../../../testdata/unittest/pip/attributes/attributes.yaml",
			want: models.NewAttributeSet(
				models.NewAttribute("maandag", 1),
				models.NewAttribute("dinsdag", 2),
				models.NewAttribute("woensdag", 3),
				models.NewAttribute("donderdag", 4),
				models.NewAttribute("vrijdag", 5),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := util.NewDummyHandler(slog.LevelDebug)

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

func TestLoad(t *testing.T) {
	testCases := []struct {
		name           string
		path1          string
		path2          string
		recurse        bool
		wantLog        int
		wantAttributes models.AttributeSet
		wantEntities   models.EntitySet
	}{
		{
			name:    "no stores",
			recurse: true,
		},
		{
			name:    "bad attr store",
			path1:   "/not/another/existing/folder",
			recurse: true,
			wantLog: 1,
		},
		{
			name:    "bad entity store",
			path1:   "",
			path2:   "/this/is/also/not/real",
			recurse: true,
			wantLog: 1,
		},
		{
			name:    "bad stores",
			path1:   "/not/another/existing/folder",
			path2:   "/this/is/also/not/real",
			recurse: true,
			wantLog: 2,
		},
		{
			name:           "attr store, no recurse",
			path1:          "../../../testdata/unittest/pip",
			wantLog:        1,
			wantAttributes: models.NewAttributeSet(),
			wantEntities:   models.NewEntitySet(),
		},
		{
			name:    "attr store, recurse",
			path1:   "../../../testdata/unittest/pip",
			recurse: true,
			wantLog: 3,
			wantAttributes: models.NewAttributeSet(
				models.NewAttribute("maandag", 1),
				models.NewAttribute("dinsdag", 2),
				models.NewAttribute("woensdag", 3),
				models.NewAttribute("donderdag", 4),
				models.NewAttribute("vrijdag", 5),
			),
			wantEntities: models.NewEntitySet(),
		},
		{
			name:           "entity store, no recurse",
			path2:          "../../../testdata/unittest/pip",
			wantLog:        1,
			wantAttributes: models.NewAttributeSet(),
			wantEntities:   models.NewEntitySet(),
		},
		{
			name:           "entity store, recurse",
			path2:          "../../../testdata/unittest/pip",
			recurse:        true,
			wantLog:        1,
			wantAttributes: models.NewAttributeSet(),
			wantEntities: models.NewEntitySet(
				models.NewEntity("App", "app1", models.NewAttributeSet(
					models.NewAttribute("code", "app1"),
					models.NewAttribute("name", "App-1"),
				)),
				models.NewEntity("App", "app2", models.NewAttributeSet(
					models.NewAttribute("code", "app2"),
					models.NewAttribute("name", "App-2"),
				)),
				models.NewEntity("App", "app3", models.NewAttributeSet(
					models.NewAttribute("code", "app3"),
					models.NewAttribute("name", "App-3"),
				), "App::app1", "App::app2",
				),
			),
		},
		{
			name:           "all, no recurse",
			path1:          "../../../testdata/unittest/pip",
			path2:          "../../../testdata/unittest/pip",
			wantLog:        2,
			wantAttributes: models.NewAttributeSet(),
			wantEntities:   models.NewEntitySet(),
		},
		{
			name:    "all, recurse",
			path1:   "../../../testdata/unittest/pip",
			path2:   "../../../testdata/unittest/pip",
			recurse: true,
			wantLog: 4,
			wantAttributes: models.NewAttributeSet(
				models.NewAttribute("maandag", 1),
				models.NewAttribute("dinsdag", 2),
				models.NewAttribute("woensdag", 3),
				models.NewAttribute("donderdag", 4),
				models.NewAttribute("vrijdag", 5),
			),
			wantEntities: models.NewEntitySet(
				models.NewEntity("App", "app1", models.NewAttributeSet(
					models.NewAttribute("code", "app1"),
					models.NewAttribute("name", "App-1"),
				)),
				models.NewEntity("App", "app2", models.NewAttributeSet(
					models.NewAttribute("code", "app2"),
					models.NewAttribute("name", "App-2"),
				)),
				models.NewEntity("App", "app3", models.NewAttributeSet(
					models.NewAttribute("code", "app3"),
					models.NewAttribute("name", "App-3"),
				), "App::app1", "App::app2",
				),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := util.NewDummyHandler(slog.LevelDebug)

			p := &pip{
				attrStore:     tc.path1,
				entityStore:   tc.path2,
				recurse:       tc.recurse,
				logger:        slog.New(h),
				attributes:    models.NewAttributeSet(),
				entities:      models.NewEntitySet(),
				newAttributes: models.NewAttributeSet,
			}

			p.loadFromStore()

			assert.Equal(t, tc.wantLog, h.Count())

			if tc.wantLog == 0 {
				if tc.wantAttributes != nil {
					tc.wantAttributes.IterateAttributes(func(key string, v1 any) {
						v2 := p.attributes.GetAttribute(key)
						assert.EqualValues(t, v1, v2)
					})

					p.attributes.IterateAttributes(func(key string, v1 any) {
						v2 := tc.wantAttributes.GetAttribute(key)
						assert.EqualValues(t, v1, v2)
					})
				}

				if tc.wantEntities != nil {
					tc.wantEntities.IterateEntities(func(e1 models.Entity) {
						e2 := p.entities.GetEntity(e1.UID())
						require.NotNil(t, e2)
						assert.EqualValues(t, e1, e2)
					})

					p.entities.IterateEntities(func(e1 models.Entity) {
						e2 := tc.wantEntities.GetEntity(e1.UID())
						require.NotNil(t, e2)
						assert.EqualValues(t, e1, e2)
					})
				}
			}
		})
	}
}
