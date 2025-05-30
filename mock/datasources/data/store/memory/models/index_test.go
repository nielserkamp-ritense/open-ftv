package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/types"
)

func TestIndexAsRecord(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		ix   *schema.Index
		want map[string]any
	}{
		{
			name: "basic",
			ix: &schema.Index{
				Parent:      schema.Parent{ID: "ix1"},
				Description: "index 1",
			},
			want: map[string]any{
				"fqdn":        "ix1",
				"id":          "ix1",
				"description": "index 1",
			},
		},
		{
			name: "fields",
			ix: &schema.Index{
				Parent:      schema.Parent{ID: "ix2"},
				Description: "index 2",
				Fields:      []string{"created", "id"},
			},
			want: map[string]any{
				"fqdn":        "ix2",
				"id":          "ix2",
				"description": "index 2",
				"fields":      []string{"created", "id"},
			},
		},
		{
			name: "fields & orders",
			ix: &schema.Index{
				Parent:      schema.Parent{ID: "ix3"},
				Description: "index 3",
				Fields:      []string{"created", "id"},
				Orders:      []types.OrderType{types.OrderAscending, types.OrderDescending},
			},
			want: map[string]any{
				"fqdn":        "ix3",
				"id":          "ix3",
				"description": "index 3",
				"fields":      []string{"created", "id"},
				"orders":      []types.OrderType{types.OrderAscending, types.OrderDescending},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := indexAsRow(tc.ix)
			assert.EqualValues(t, tc.want, got.Data)
		})
	}
}
