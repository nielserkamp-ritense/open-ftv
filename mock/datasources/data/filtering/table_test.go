package filtering

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

func TestBuildTableFilter(t *testing.T) {
	t.Parallel()

	t1 := &schema.Table{
		Object: schema.Object{
			Parent:      schema.Parent{ID: "t1"},
			Description: "table 1",
			Fields: []*schema.Field{
				{Object: schema.Object{Parent: schema.Parent{ID: "f1"}, Description: "field 1"}, Type: enums.IntegerType},
				{Object: schema.Object{Parent: schema.Parent{ID: "f2"}, Description: "field 2"}, Type: enums.StringType},
				{Object: schema.Object{Parent: schema.Parent{ID: "f3"}, Description: "field 3"}, Type: enums.FloatType},
			},
		},
	}

	testCases := []struct {
		name   string
		t      *schema.Table
		filter map[string]any
		want   map[string]any
	}{
		{name: "no filter", t: t1},
		{name: "empty filter", t: t1, filter: map[string]any{}, want: map[string]any{}},
		{name: "no matching fields", t: t1, filter: map[string]any{"f4": "hello", "f5": 123, "f6": true}, want: map[string]any{}},
		{name: "one matching field", t: t1, filter: map[string]any{"f4": "hello", "t1.f1": 123, "f6": true}, want: map[string]any{"f1": 123}},
		{name: "all matching fields", t: t1, filter: map[string]any{"t1.f2": "hello", "f1": 123, "t1.f3": 1.25}, want: map[string]any{"f2": "hello", "f1": 123, "f3": 1.25}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewTableFilter(tc.t, tc.filter)
			assert.EqualValues(t, tc.want, got)
		})
	}
}
