package pip

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestValidPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		path string
		want bool
	}{
		{name: "empty"},
		{name: "dot", path: ".", want: true},
		{name: "dot dot", path: "..", want: true},
		{name: "invalid", path: "/not/really/a/valid/path"},
		{name: "valid", path: "/usr/sbin", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := validPath(tc.path)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		level          slog.Level
		path           string
		recurse        bool
		wantLog        int
		wantAttributes models.AttributeSet
		wantEntities   models.EntitySet
	}{
		{
			name:         "no store",
			recurse:      true,
			wantLog:      1,
			wantEntities: models.NewEntitySet(),
		},
		{
			name:         "invalid store",
			path:         "/not/a/valid/path",
			recurse:      true,
			wantLog:      1,
			wantEntities: models.NewEntitySet(),
		},
		{
			name:    "with store, no recurse",
			level:   slog.LevelDebug,
			path:    "../../../testdata/unittest/pip",
			wantLog: 1,
			wantAttributes: models.NewAttributeSet(
				models.NewOriginalAttribute("maandag", uint64(1), uint64(1), ""),
				models.NewOriginalAttribute("dinsdag", uint64(2), uint64(2), ""),
				models.NewOriginalAttribute("woensdag", uint64(3), uint64(3), ""),
				models.NewOriginalAttribute("donderdag", uint64(4), uint64(4), ""),
				models.NewOriginalAttribute("vrijdag", uint64(5), uint64(5), ""),
			),
			wantEntities: models.NewEntitySet(
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

			h := util.NewDummyHandler(tc.level)
			logger := slog.New(h)

			p1 := New(nil, logger, WithFileStore(tc.path, tc.recurse))
			require.NotNil(t, p1)

			p2, ok := p1.(*pip)
			require.True(t, ok)
			require.NotNil(t, p2)

			assert.Equal(t, tc.wantLog, h.Count())

			if tc.wantAttributes != nil {
				list, err := p2.attributePersist.List()
				require.NoError(t, err)

				for i := range list {
					attr := list[i]
					attr2 := tc.wantAttributes.GetAttribute(attr.Key())
					require.NotNil(t, attr2)
					assert.True(t, models.AttributeEqual(attr, attr2))
				}
			}

			if tc.wantEntities != nil {
				tc.wantEntities.IterateEntities(func(e1 models.Entity) {
					e2 := p2.GetEntity(e1.UID())
					require.NotNil(t, e2)
					assert.True(t, models.EntityEqual(e1, e2))
				})

				p2.IterateEntities(func(e1 models.Entity) {
					e2 := tc.wantEntities.GetEntity(e1.UID())
					require.NotNil(t, e2)
					assert.True(t, models.EntityEqual(e1, e2))
				})
			}
		})
	}
}

func TestPIP_Attributes(t *testing.T) {
	t.Parallel()

	t.Run("pip as AttributeSet", func(t *testing.T) {
		h := util.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		p := New(context.Background(), logger)
		require.NotNil(t, p)

		p.AddAttribute("hello", "world")
		p.AddAttribute("int", "987")
		p.AddAttribute("float", "987.789")

		assert.Equal(t, "world", p.GetAttributeValue("hello"))
		assert.Nil(t, p.GetAttribute("bool"))

		// p2 := &pip{attributes: models.NewAttributeSet(models.NewAttribute("hello", "world2"), models.NewAttribute("bool", true))}

		p2 := New(context.Background(), logger)
		p2.AddAttribute("hello", "world2")
		p2.AddAttribute("bool", true)

		p.MergeAttributes(p2)

		assert.Equal(t, "world2", p.GetAttributeValue("hello"))
		assert.Equal(t, true, p.GetAttributeValue("bool"))
		assert.Equal(t, "987", p.GetAttributeValue("int"))

		p.RemoveAttribute("bool")
		p.RemoveAttribute("int")
		assert.Nil(t, p.GetAttributeValue("bool"))
		assert.Nil(t, p.GetAttributeValue("int"))

		var count int
		p.IterateAttributes(func(models.Attribute) {
			count++
		})
		assert.Equal(t, 2, count)
	})
}

func TestPIP_Entities(t *testing.T) {
	t.Parallel()

	t.Run("pip as EntitySet", func(t *testing.T) {
		h := util.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		p := New(nil, logger).(*pip)
		require.NotNil(t, p)

		p.AddEntity(models.NewEntity("x", "y", models.NewAttributeSet()))
		p.AddEntity(models.NewEntity("x", "z", models.NewAttributeSet()))

		p.MergeEntities(
			models.NewEntitySet(
				models.NewEntity("q", "x", models.NewAttributeSet()),
				models.NewEntity("q", "y", models.NewAttributeSet()),
				models.NewEntity("q", "z", models.NewAttributeSet()),
			),
		)

		var count int
		p.IterateEntities(func(entity models.Entity) {
			count++
		})
		assert.Equal(t, 5, count)

		e := p.GetEntity("x::y")
		require.NotNil(t, e)

		e = p.GetEntity("x::x")
		require.Nil(t, e)

		p.RemoveEntity("q::y")
		p.RemoveEntity("q::x")

		count = 0
		p.IterateEntities(func(entity models.Entity) {
			count++
		})
		assert.Equal(t, 3, count)

		e = p.GetEntity("q::y")
		require.Nil(t, e)
	})
}
