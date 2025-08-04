package pip

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestPIP_Attributes(t *testing.T) {
	t.Parallel()

	t.Run("pip as AttributeSet", func(t *testing.T) {
		t.Parallel()

		h := util.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		p := New(context.Background(), logger)
		require.NotNil(t, p)

		p.AddAttribute("hello", "world")
		p.AddAttributeWithType("int", "987", "xsd:string")
		p.AddOriginalAttribute("float", 987.789, "987.789", "xsd:number")

		assert.Equal(t, "world", p.GetAttributeValue("hello"))
		assert.Nil(t, p.GetAttribute("bool"))

		p2 := models.NewAttributeSet()
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

			h := util.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			p := New(context.Background(), logger)
			require.NotNil(t, p)

			eh := &myAttributeSink{}
			p.AddEventSink(eh)

			p.AddAttribute("hello", "world")
			p.AddAttribute("int", "987")
			p.AddAttribute("float", "987.789")

			assert.Equal(t, 3, eh.inserts)
			assert.Zero(t, eh.updates)
			assert.Zero(t, eh.deletes)
			eh.inserts = 0

			p.ReplaceAllAttributes(tc.list)

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
