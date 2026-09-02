package pip

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestPIP_Attributes(t *testing.T) {
	t.Parallel()

	t.Run("pip as AttributeSet", func(t *testing.T) {
		t.Parallel()

		h := util.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		p, err := New(context.Background(), logger, WithKeyValueDB(memory.New(), ""))
		require.NoError(t, err)
		require.NotNil(t, p)

		a1, err1 := p.AddAttributeKV("hello", "world")
		require.NoError(t, err1)
		require.NotNil(t, a1)

		a2, err2 := p.AddAttributeKVWithType("int", "987", "xsd:string")
		require.NoError(t, err2)
		require.NotNil(t, a2)

		a3, err3 := p.AddOriginalAttribute("float", 987.789, "987.789", "xsd:number")
		require.NoError(t, err3)
		require.NotNil(t, a3)

		assert.Equal(t, "world", p.GetAttributeValue("hello"))

		a4, _, err4 := p.GetAttribute("bool")
		require.NoError(t, err4)
		assert.Nil(t, a4)

		p2 := models.NewAttributeSet()
		p2.AddAttributeKV("hello", "world2")
		p2.AddAttributeKV("bool", true)

		p.MergeAttributes(p2)

		var count int
		p.IterateAttributes(func(*models.Attribute) {
			count++
		})
		assert.Equal(t, 4, count)

		assert.Equal(t, "world2", p.GetAttributeValue("hello"))
		assert.Equal(t, true, p.GetAttributeValue("bool"))
		assert.Equal(t, "987", p.GetAttributeValue("int"))

		a1, err1 = p.RemoveAttribute("bool", identity.NewSystemPrincipal())
		require.NoError(t, err1)
		require.NotNil(t, a1)

		a2, err2 = p.RemoveAttribute("int", identity.NewSystemPrincipal())
		require.NoError(t, err2)
		require.NotNil(t, a2)

		assert.Nil(t, p.GetAttributeValue("bool"))
		assert.Nil(t, p.GetAttributeValue("int"))

		count = 0
		p.IterateAttributes(func(*models.Attribute) {
			count++
		})
		assert.Equal(t, 2, count)
	})
}

func TestPIP_ReplaceAllAttributes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		list *models.AttributeSet
		want int
	}{
		{
			name: "nil",
		},
		{
			name: "empty",
			list: models.NewAttributeSet(),
		},
		{
			name: "one",
			list: models.NewAttributeSet(
				models.NewAttribute("x", "y"),
			),
			want: 1,
		},
		{
			name: "three",
			list: models.NewAttributeSet(
				models.NewAttribute("hello", "world"),
				models.NewAttribute("x", "y"),
				models.NewAttribute("int", "1"),
			),
			want: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := util.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			p, err := New(ctx, logger, WithKeyValueDB(memory.New(), ""))
			require.NoError(t, err)
			require.NotNil(t, p)

			eh := &myAttributeSink{}
			p.AddEventSink(eh)

			_, _ = p.AddAttributeKV("hello", "world")
			_, _ = p.AddAttributeKV("int", "987")
			_, _ = p.AddAttributeKV("float", "987.789")

			assert.Equal(t, 3, eh.inserts)
			assert.Zero(t, eh.updates)
			assert.Zero(t, eh.deletes)
			eh.inserts = 0

			p.ReplaceAllAttributes(tc.list, identity.NewSystemPrincipal())

			assert.Equal(t, tc.want, eh.inserts)
			assert.Zero(t, eh.updates)
			assert.Equal(t, 3, eh.deletes)

			var count int
			p.IterateAttributes(func(*models.Attribute) {
				count++
			})
			assert.Equal(t, tc.want, count)
		})
	}
}

type myAttributeSink struct {
	inserts int
	updates int
	deletes int
}

func (s *myAttributeSink) Handle(t models.EventType, _ string) {
	switch t {
	case models.AttributeAdded:
		s.inserts++
	case models.AttributeReplaced:
		s.updates++
	case models.AttributeRemoved:
		s.deletes++
	default:
	}
}
