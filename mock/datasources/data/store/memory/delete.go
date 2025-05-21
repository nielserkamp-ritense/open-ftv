package memory

import "fmt"

// DeleteRecord implements the Maintainer interface.
func (s *storage) DeleteRecord(tableID string, pk []any) error {

	return fmt.Errorf("not implemented")
}
