package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

func TestForeignKeyAsRecord(t *testing.T) {
	t.Parallel()

	t.Run("ForeignKey as record", func(t *testing.T) {
		def := &schema.ForeignKey{
			Index: schema.Index{
				Parent:      schema.Parent{ID: "fk1"},
				Description: "foreign key 1",
				Fields:      []string{"created", "id"},
			},
			ForeignTable: "foreign1",
		}

		got := fkAsRow(def)
		require.NotNil(t, got)

		assert.Equal(t, "fk1", got.Data["fqdn"])
		assert.Equal(t, "fk1", got.Data["id"])
		assert.Equal(t, "foreign key 1", got.Data["description"])
		assert.Equal(t, "foreign1", got.Data["table"])
		assert.EqualValues(t, []string{"created", "id"}, got.Data["fields"])
	})
}
