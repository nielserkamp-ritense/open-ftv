package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObject_IterateFields(t *testing.T) {
	t.Parallel()

	t.Run("object iterate fields", func(t *testing.T) {
		t.Parallel()

		o := &Object{
			Parent:      Parent{ID: "o1"},
			Description: "object 1",
			Fields: []*Field{
				{Object: Object{Parent: Parent{ID: "f1"}}},
				{Object: Object{Parent: Parent{ID: "f2"}}},
				{Object: Object{Parent: Parent{ID: "f3"}}},
				{Object: Object{Parent: Parent{ID: "f4"}}},
				{Object: Object{Parent: Parent{ID: "f5"}}},
			},
		}

		var count int
		o.IterateFields(func(field *Field) {
			count++
			require.NotNil(t, field)
		})

		assert.Equal(t, 5, count)
	})
}
