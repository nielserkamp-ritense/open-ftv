package pip

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
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
		wantAttributes *models.AttributeSet
		wantEntities   *models.EntitySet
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
			path:    "../../testdata/unittest/pip",
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

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := util.NewDummyHandler(tc.level)
			logger := slog.New(h)

			p1, err := New(t.Context(), logger, WithKeyValueDB(memory.New(), ""), WithFileStore(tc.path, tc.recurse))
			require.NoError(t, err)
			require.NotNil(t, p1)

			assert.Equal(t, tc.wantLog, h.Count())

			if tc.wantAttributes != nil {
				list, err := p1.attributeDB.ListAttributes(ctx)
				require.NoError(t, err)

				for i := range list {
					attr := list[i]
					attr2 := tc.wantAttributes.GetAttribute(attr.Key())
					require.NotNil(t, attr2)
					assert.True(t, attr.Equals(attr2))
				}
			}

			if tc.wantEntities != nil {
				tc.wantEntities.IterateEntities(func(e1 *models.Entity) {
					e2, _, err2 := p1.GetEntity(e1.UID())
					require.NoError(t, err2)
					require.NotNil(t, e2)
					assert.True(t, e1.Equals(e2))
				})

				p1.IterateEntities(func(e1 *models.Entity) {
					e2 := tc.wantEntities.GetEntity(e1.UID())
					require.NotNil(t, e2)
					assert.True(t, e1.Equals(e2))
				})
			}
		})
	}
}
