package pip

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/storage/valkeyrie/memory"
)

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
			wantLog: 1,
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
			wantLog: 2,
			wantAttributes: models.NewAttributeSet(
				models.NewAttribute("maandag", 1),
				models.NewAttribute("dinsdag", 2),
				models.NewAttribute("woensdag", 3),
				models.NewAttribute("donderdag", 4),
				models.NewAttribute("vrijdag", 5),
			),
			wantEntities: models.NewEntitySet(
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

			s := memory.New()
			ap := NewAttributeStore(context.Background(), s, "")

			p := &pip{
				attrStore:        tc.path1,
				entityStore:      tc.path2,
				recurse:          tc.recurse,
				logger:           slog.New(h),
				entities:         models.NewEntitySet(),
				newAttributes:    models.NewAttributeSet,
				store:            s,
				attributePersist: ap,
			}

			p.loadFromStore()

			assert.Equal(t, tc.wantLog, h.Count())

			if tc.wantLog == 0 {
				if tc.wantAttributes != nil {
					tc.wantAttributes.IterateAttributes(func(attr models.Attribute) {
						v2 := p.GetAttribute(attr.Key())
						assert.EqualValues(t, attr.Value(), v2)
					})

					p.IterateAttributes(func(attr models.Attribute) {
						v2 := tc.wantAttributes.GetAttribute(attr.Key())
						assert.EqualValues(t, attr.Value(), v2)
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
