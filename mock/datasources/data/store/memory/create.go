package memory

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/memory/models"
)

// CreateRecord implements the Maintainer interface.
func (s *storage) CreateRecord(tableID string, record *models.Row) error {
	table, err := s.GetTable(tableID)
	if err != nil {
		return fmt.Errorf("createRecord: %w", err)
	}

	table.CreateRow(record)
	return nil
}
