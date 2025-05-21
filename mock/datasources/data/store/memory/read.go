package memory

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/memory/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/types"
)

// SelectPK implements the Reader interface.
func (s *storage) SelectPK(tableID string, pk []any) (*models.Row, error) {
	table, err := s.GetTable(tableID)
	if err != nil {
		return nil, fmt.Errorf("selectPK: %w", err)
	}

	if len(table.Definition().PrimaryKey.Fields) != len(pk) {
		return nil, fmt.Errorf("selectPK: field count mismatch")
	}

	t, err2 := s.findUnqualifiedTable(table.Definition().ID)
	if err2 != nil {
		return nil, fmt.Errorf("selectPK: %w", err2)
	}

	key := models.KeyFromData(pk, table.Definition().PrimaryKey)
	if rec := t.PK[key]; rec != nil {
		return rec, nil
	}
	return nil, fmt.Errorf("selectPK: primary key %v not found", pk)
}

// SelectIX implements the Reader interface.
func (s *storage) SelectIX(tableID string, id string, keys []any) (models.Rows, error) {
	table, err := s.GetTable(tableID)
	if err != nil {
		return nil, fmt.Errorf("selectIX: %w", err)
	}

	index := table.Definition().SecondaryIndex(id)
	if index == nil {
		return nil, fmt.Errorf("selectIX: index [%s] not found", id)
	}

	if len(index.Fields) != len(keys) {
		return nil, fmt.Errorf("selectIX: field count mismatch")
	}

	t, err2 := s.findUnqualifiedTable(table.Definition().ID)
	if err2 != nil {
		return nil, fmt.Errorf("selectIX: %w", err2)
	}

	key := models.KeyFromData(keys, index)
	if indexData := t.Indexes[id]; indexData != nil {
		if list := indexData[key]; len(list) > 0 {
			return list, nil
		}
	}
	return nil, fmt.Errorf("selectIX: keys %v in index [%s] not found", keys, id)
}

// Search implements the Reader interface.
func (s *storage) Search(tableID string, filter map[string]any) (models.Rows, error) {
	table, err := s.GetTable(tableID)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	t, err2 := s.findUnqualifiedTable(table.Definition().ID)
	if err2 != nil {
		return nil, fmt.Errorf("search: %w", err2)
	}

	expr, _ := filter["fields"].(string)
	delete(filter, "fields")
	matcher := types.NewFieldMatcher(expr)

	var out models.Rows
	for i := range t.Data {
		rec := t.Data[i]
		if rec.MatchFilter(filter) {
			out = append(out, rec.MatchFields(matcher))
		}
	}

	if len(out) > 0 {
		return out, nil
	}
	return nil, fmt.Errorf("search: no matching records found for {%v}", filter)
}
