package pip

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/standards"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestLoadEntities(t *testing.T) {
	testCases := []struct {
		name    string
		path    string
		wantLog int
		want    standards.EntitySet
	}{
		{
			name:    "empty",
			wantLog: 1,
		},
		{
			name:    "bad path",
			path:    "/this/is/another/bad/file",
			wantLog: 1,
			want:    standards.NewEntitySet(),
		},
		{
			name:    "not yaml",
			path:    "../../../testdata/unittest/pip/not_yaml.yaml",
			wantLog: 1,
			want:    standards.NewEntitySet(),
		},
		{
			name: "1 entity",
			path: "../../../testdata/unittest/pip/entities/entity.yaml",
			want: standards.NewEntitySet(
				standards.NewEntity("app", "app1", standards.NewAttributeSet(
					standards.NewAttribute("code", "app1"),
					standards.NewAttribute("name", "App-1"),
				)),
			),
		},
		{
			name: "few entities",
			path: "../../../testdata/unittest/pip/entities/entities.yaml",
			want: standards.NewEntitySet(
				standards.NewEntity("app", "app1", standards.NewAttributeSet(
					standards.NewAttribute("code", "app1"),
					standards.NewAttribute("name", "App-1"),
				)),
				standards.NewEntity("app", "app2", standards.NewAttributeSet(
					standards.NewAttribute("code", "app2"),
					standards.NewAttribute("name", "App-2"),
				)),
				standards.NewEntity("app", "app3", standards.NewAttributeSet(
					standards.NewAttribute("code", "app3"),
					standards.NewAttribute("name", "App-3"),
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
				entities:      standards.NewEntitySet(),
				newAttributes: standards.NewAttributeSet,
			}

			p.loadEntities(tc.path)

			assert.Equal(t, tc.wantLog, h.Count())

			if tc.wantLog == 0 {
				tc.want.IterateEntities(func(e1 standards.Entity) {
					e2 := p.entities.GetEntity(e1.UID())
					require.NotNil(t, e2)
					assert.EqualValues(t, e1, e2)
				})

				p.entities.IterateEntities(func(e1 standards.Entity) {
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
		want    standards.AttributeSet
	}{
		{
			name:    "empty",
			wantLog: 1,
		},
		{
			name:    "bad path",
			path:    "/this/is/another/bad/file",
			wantLog: 1,
			want:    standards.NewAttributeSet(),
		},
		{
			name:    "not yaml",
			path:    "../../../testdata/unittest/pip/not_yaml.yaml",
			wantLog: 1,
			want:    standards.NewAttributeSet(),
		},
		{
			name: "1 attribute",
			path: "../../../testdata/unittest/pip/attributes/attribute.yaml",
			want: standards.NewAttributeSet(
				standards.NewAttribute("maandag", 1),
			),
		},
		{
			name: "few attributes",
			path: "../../../testdata/unittest/pip/attributes/attributes.yaml",
			want: standards.NewAttributeSet(
				standards.NewAttribute("maandag", 1),
				standards.NewAttribute("dinsdag", 2),
				standards.NewAttribute("woensdag", 3),
				standards.NewAttribute("donderdag", 4),
				standards.NewAttribute("vrijdag", 5),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := util.NewDummyHandler(slog.LevelDebug)

			p := &pip{
				logger:     slog.New(h),
				attributes: standards.NewAttributeSet(),
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
		wantAttributes standards.AttributeSet
		wantEntities   standards.EntitySet
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
			wantAttributes: standards.NewAttributeSet(),
			wantEntities:   standards.NewEntitySet(),
		},
		{
			name:    "attr store, recurse",
			path1:   "../../../testdata/unittest/pip",
			recurse: true,
			wantLog: 1,
			wantAttributes: standards.NewAttributeSet(
				standards.NewAttribute("maandag", 1),
				standards.NewAttribute("dinsdag", 2),
				standards.NewAttribute("woensdag", 3),
				standards.NewAttribute("donderdag", 4),
				standards.NewAttribute("vrijdag", 5),
			),
			wantEntities: standards.NewEntitySet(),
		},
		{
			name:           "entity store, no recurse",
			path2:          "../../../testdata/unittest/pip",
			wantLog:        1,
			wantAttributes: standards.NewAttributeSet(),
			wantEntities:   standards.NewEntitySet(),
		},
		{
			name:           "entity store, recurse",
			path2:          "../../../testdata/unittest/pip",
			recurse:        true,
			wantLog:        1,
			wantAttributes: standards.NewAttributeSet(),
			wantEntities: standards.NewEntitySet(
				standards.NewEntity("App", "app1", standards.NewAttributeSet(
					standards.NewAttribute("code", "app1"),
					standards.NewAttribute("name", "App-1"),
				)),
				standards.NewEntity("App", "app2", standards.NewAttributeSet(
					standards.NewAttribute("code", "app2"),
					standards.NewAttribute("name", "App-2"),
				)),
				standards.NewEntity("App", "app3", standards.NewAttributeSet(
					standards.NewAttribute("code", "app3"),
					standards.NewAttribute("name", "App-3"),
				), "App::app1", "App::app2",
				),
			),
		},
		{
			name:           "all, no recurse",
			path1:          "../../../testdata/unittest/pip",
			path2:          "../../../testdata/unittest/pip",
			wantLog:        2,
			wantAttributes: standards.NewAttributeSet(),
			wantEntities:   standards.NewEntitySet(),
		},
		{
			name:    "all, recurse",
			path1:   "../../../testdata/unittest/pip",
			path2:   "../../../testdata/unittest/pip",
			recurse: true,
			wantLog: 2,
			wantAttributes: standards.NewAttributeSet(
				standards.NewAttribute("maandag", 1),
				standards.NewAttribute("dinsdag", 2),
				standards.NewAttribute("woensdag", 3),
				standards.NewAttribute("donderdag", 4),
				standards.NewAttribute("vrijdag", 5),
			),
			wantEntities: standards.NewEntitySet(
				standards.NewEntity("App", "app1", standards.NewAttributeSet(
					standards.NewAttribute("code", "app1"),
					standards.NewAttribute("name", "App-1"),
				)),
				standards.NewEntity("App", "app2", standards.NewAttributeSet(
					standards.NewAttribute("code", "app2"),
					standards.NewAttribute("name", "App-2"),
				)),
				standards.NewEntity("App", "app3", standards.NewAttributeSet(
					standards.NewAttribute("code", "app3"),
					standards.NewAttribute("name", "App-3"),
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
				attributes:    standards.NewAttributeSet(),
				entities:      standards.NewEntitySet(),
				newAttributes: standards.NewAttributeSet,
			}

			p.load()

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
					tc.wantEntities.IterateEntities(func(e1 standards.Entity) {
						e2 := p.entities.GetEntity(e1.UID())
						require.NotNil(t, e2)
						assert.EqualValues(t, e1, e2)
					})

					p.entities.IterateEntities(func(e1 standards.Entity) {
						e2 := tc.wantEntities.GetEntity(e1.UID())
						require.NotNil(t, e2)
						assert.EqualValues(t, e1, e2)
					})
				}
			}
		})
	}
}
