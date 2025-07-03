package pep

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func TestDetermineAction(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		action models.Entity
		req    *models.HTTPRequest
		want   models.Entity
	}{
		{
			name:   "empty, no method",
			action: models.NewEntity("", "", models.NewAttributeSet()),
			req:    &models.HTTPRequest{},
			want:   models.NewEntity("", "", models.NewAttributeSet()),
		},
		{
			name:   "empty, GET",
			action: models.NewEntity("", "", models.NewAttributeSet()),
			req:    &models.HTTPRequest{Method: "GET"},
			want:   models.NewEntity("name", "can_read", models.NewAttributeSet(models.NewAttribute("method", "GET"))),
		},
		{
			name:   "empty, POST",
			action: models.NewEntity("", "", models.NewAttributeSet()),
			req:    &models.HTTPRequest{Method: "POST"},
			want:   models.NewEntity("name", "can_update", models.NewAttributeSet(models.NewAttribute("method", "POST"))),
		},
		{
			name:   "empty, PUT",
			action: models.NewEntity("", "", models.NewAttributeSet()),
			req:    &models.HTTPRequest{Method: "PUT"},
			want:   models.NewEntity("name", "can_create", models.NewAttributeSet(models.NewAttribute("method", "PUT"))),
		},
		{
			name:   "empty, DELETE",
			action: models.NewEntity("", "", models.NewAttributeSet()),
			req:    &models.HTTPRequest{Method: "DELETE"},
			want:   models.NewEntity("name", "can_delete", models.NewAttributeSet(models.NewAttribute("method", "DELETE"))),
		},
		{
			name:   "empty, OPTIONS",
			action: models.NewEntity("", "", models.NewAttributeSet()),
			req:    &models.HTTPRequest{Method: "OPTIONS"},
			want:   models.NewEntity("name", "can_read", models.NewAttributeSet(models.NewAttribute("method", "OPTIONS"))),
		},
		{
			name:   "not empty, no method",
			action: models.NewEntity("action", "read", models.NewAttributeSet()),
			req:    &models.HTTPRequest{},
			want:   models.NewEntity("action", "read", models.NewAttributeSet()),
		},
		{
			name:   "not empty, GET",
			action: models.NewEntity("action", "read", models.NewAttributeSet()),
			req:    &models.HTTPRequest{Method: "GET"},
			want:   models.NewEntity("action", "read", models.NewAttributeSet(models.NewAttribute("method", "GET"))),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := &collector{
				req:  tc.req,
				parc: &models.PARC{Action: tc.action},
			}

			c.determineAction()

			assert.Equal(t, tc.want.Type(), c.parc.Action.Type())
			assert.Equal(t, tc.want.ID(), c.parc.Action.ID())

			tc.want.Attributes().IterateAttributes(func(a1 models.Attribute) {
				a2 := c.parc.Action.Attributes().GetAttribute(a1.Key())
				assert.EqualValues(t, a1, a2)
			})

			c.parc.Action.Attributes().IterateAttributes(func(a1 models.Attribute) {
				a2 := tc.want.Attributes().GetAttribute(a1.Key())
				assert.EqualValues(t, a1, a2)
			})
		})
	}
}
