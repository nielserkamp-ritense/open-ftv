package memory

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store/memory/models"
)

// UpdateRecord implements the Maintainer interface.
func (s *storage) UpdateRecord(tableID string, pk []any, record *models.Row) error {
	table, err := s.GetTable(tableID)
	if err != nil {
		return fmt.Errorf("updateRecord: %w", err)
	}

	// We call the delete and create function to emulate an update.
	// This simplifies the code and the extra cost is minimal.

	if err = table.DeleteRow(pk); err != nil {
		return err
	}

	return table.CreateRow(record)
}
