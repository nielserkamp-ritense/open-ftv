package memory

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store/memory/models"
)

// UpdateRecord implements the Maintainer interface.
func (s *storage) UpdateRecord(tableID string, pk []any, record *models.Row) error {

	return fmt.Errorf("not implemented")
}
