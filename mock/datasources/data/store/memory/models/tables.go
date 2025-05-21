package models

// Tables is a convenience type for a list of tables.
type Tables []*Table

func (t Tables) MarshalCSV() ([]byte, error) {
	return nil, nil
}
