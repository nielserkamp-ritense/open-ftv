package pip

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestPIP_Entities(t *testing.T) {
	t.Parallel()

	t.Run("pip as EntitySet", func(t *testing.T) {
		h := util.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		p := New(nil, logger)
		require.NotNil(t, p)

		_, _ = p.AddEntity(models.NewEntity("x", "y", models.NewAttributeSet()))
		_, _ = p.AddEntity(models.NewEntity("x", "z", models.NewAttributeSet()))

		p.MergeEntities(
			models.NewEntitySet(
				models.NewEntity("q", "x", models.NewAttributeSet()),
				models.NewEntity("q", "y", models.NewAttributeSet()),
				models.NewEntity("q", "z", models.NewAttributeSet()),
			),
		)

		var count int
		p.IterateEntities(func(entity *models.Entity) {
			count++
		})
		assert.Equal(t, 5, count)

		e, _, err := p.GetEntity("x::y")
		require.NoError(t, err)
		require.NotNil(t, e)

		e, _, err = p.GetEntity("x::x")
		require.NoError(t, err)
		require.Nil(t, e)

		_, _ = p.RemoveEntity("q::y", identity.NewSystemPrincipal())
		_, _ = p.RemoveEntity("q::x", identity.NewSystemPrincipal())

		count = 0
		p.IterateEntities(func(entity *models.Entity) {
			count++
		})
		assert.Equal(t, 3, count)

		e, _, err = p.GetEntity("q::y")
		require.NoError(t, err)
		require.Nil(t, e)
	})
}

func TestPIP_ReplaceAllEntities(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		list *models.EntitySet
		want int
	}{
		{
			name: "nil",
		},
		{
			name: "empty",
			list: models.NewEntitySet(),
		},
		{
			name: "one",
			list: models.NewEntitySet(
				models.NewEntity("x", "y", models.NewAttributeSet()),
			),
			want: 1,
		},
		{
			name: "three",
			list: models.NewEntitySet(
				models.NewEntity("q", "x", models.NewAttributeSet()),
				models.NewEntity("q", "y", models.NewAttributeSet()),
				models.NewEntity("z", "a", models.NewAttributeSet()),
			),
			want: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := util.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			p := New(nil, logger)
			require.NotNil(t, p)

			_, _ = p.AddEntity(models.NewEntity("x", "y", models.NewAttributeSet()))
			_, _ = p.AddEntity(models.NewEntity("x", "z", models.NewAttributeSet()))

			p.MergeEntities(
				models.NewEntitySet(
					models.NewEntity("q", "x", models.NewAttributeSet()),
					models.NewEntity("q", "y", models.NewAttributeSet()),
					models.NewEntity("q", "z", models.NewAttributeSet()),
				),
			)

			p.ReplaceAllEntities(tc.list, identity.NewSystemPrincipal())

			var count int
			p.IterateEntities(func(_ *models.Entity) {
				count++
			})
			assert.Equal(t, tc.want, count)
		})
	}
}
