package memory

import "fmt"

// DeleteRecord implements the Maintainer interface.
func (s *storage) DeleteRecord(tableID string, pk []any) error {
	table, err := s.GetTable(tableID)
	if err != nil {
		return fmt.Errorf("deleteRecord: %w", err)
	}
	return table.DeleteRow(pk)
}
