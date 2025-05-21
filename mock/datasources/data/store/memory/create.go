package memory

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/memory/models"
)

// CreateRecord implements the Maintainer interface.
func (s *storage) CreateRecord(tableID string, record *models.Row) error {
	table, err := s.GetTable(tableID)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	tableData, err2 := s.findUnqualifiedTable(table.Definition().ID)
	if err2 != nil {
		return fmt.Errorf("search: %w", err2)
	}

	tableData.CreateRecord(record)
	return nil
}
