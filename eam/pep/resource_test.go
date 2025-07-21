package pep

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func TestDetermineResource(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		resource *models.Entity
		uri      string
		want     *models.Entity
	}{
		{
			name:     "empty, no uri",
			resource: models.NewEntity("", "", models.NewAttributeSet()),
			want:     models.NewEntity("", "", models.NewAttributeSet()),
		},
		{
			name:     "empty, uri",
			resource: models.NewEntity("", "", models.NewAttributeSet()),
			uri:      "http://localhost",
			want:     models.NewEntity("service", "http://localhost", models.NewAttributeSet()),
		},
		{
			name:     "not empty, no uri",
			resource: models.NewEntity("book", "12345", models.NewAttributeSet()),
			want:     models.NewEntity("book", "12345", models.NewAttributeSet()),
		},
		{
			name:     "not empty, uri",
			resource: models.NewEntity("book", "12345", models.NewAttributeSet()),
			uri:      "http://localhost",
			want:     models.NewEntity("book", "12345", models.NewAttributeSet()),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := &collector{
				parc:   &models.PARC{Resource: tc.resource},
				newURI: tc.uri,
			}

			c.determineResource()

			assert.Equal(t, tc.want.Type(), c.parc.Resource.Type())
			assert.Equal(t, tc.want.ID(), c.parc.Resource.ID())

			tc.want.Attributes().IterateAttributes(func(a1 *models.Attribute) {
				a2 := c.parc.Resource.Attributes().GetAttribute(a1.Key())
				require.NotNil(t, a2)
				assert.EqualValues(t, a1, a2)
			})

			c.parc.Resource.Attributes().IterateAttributes(func(a1 *models.Attribute) {
				a2 := tc.want.Attributes().GetAttribute(a1.Key())
				require.NotNil(t, a2)
				assert.EqualValues(t, a1, a2)
			})
		})
	}
}
