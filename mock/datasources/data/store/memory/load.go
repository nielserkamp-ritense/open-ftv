package memory

import (
	"strings"
)

// AddTableFromData implements the Storage interface.
func (s *storage) AddTableFromData(sourceID, tableID string, data []map[string]any) error {
	tableDef, err := s.findTableDef(sourceID, tableID)
	if err != nil {
		return err
	}

	source := s.sources[strings.ToLower(sourceID)]
	source.AddTableFromData(tableDef, data)

	s.tables[tableDef.FQID()] = source.Tables[tableID]
	return nil
}

// AddTableFromCSV implements the Storage interface.
func (s *storage) AddTableFromCSV(sourceID, tableID string, csv [][]string) error {
	tableDef, err := s.findTableDef(sourceID, tableID)
	if err != nil {
		return err
	}

	source := s.sources[strings.ToLower(sourceID)]
	source.AddTableFromCSV(tableDef, csv)

	s.tables[tableDef.FQID()] = source.Tables[tableID]
	return nil
}
