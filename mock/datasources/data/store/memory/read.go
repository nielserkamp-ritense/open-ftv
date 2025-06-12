package memory

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/filtering"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/matching"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/memory/joins"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/memory/models"
)

// SelectPK implements the Reader interface.
func (s *storage) SelectPK(tableID string, pk []any, fields matching.FieldMatcher) (*models.Row, error) {
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
		return rec.MatchFields(fields), nil
	}
	return nil, fmt.Errorf("selectPK: primary key %v not found", pk)
}

// SelectIX implements the Reader interface.
func (s *storage) SelectIX(tableID string, id string, keys []any, fields matching.FieldMatcher) (models.Rows, error) {
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
			return list.MatchFields(fields), nil
		}
	}
	return nil, fmt.Errorf("selectIX: keys %v in index [%s] not found", keys, id)
}

// Search implements the Reader interface.
func (s *storage) Search(tableID string, filter filtering.Filterer, fields matching.FieldMatcher) (models.Rows, error) {
	table, err := s.GetTable(tableID)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	def := table.Definition()
	t, err2 := s.findTable(def.Parent.ID, def.ID)
	if err2 != nil {
		return nil, fmt.Errorf("search: %w", err2)
	}

	var out models.Rows
	for i := range t.Data {
		rec := t.Data[i]
		// vertical data-minimalization.
		if rec.MatchPrimary(filter) {
			// horizontal data-minimalization.
			out = append(out, table.AddTransformations(rec).MatchFields(fields))
		}
	}

	if len(out) > 0 {
		return out, nil
	}
	return nil, fmt.Errorf("search: no matching records found for {%v}", filter)
}

// GetEndpoint implements the Reader interface.
func (s *storage) GetEndpoint(e *schema.Endpoint, filter filtering.Filterer, fields matching.FieldMatcher) (models.Rows, error) {
	if e.Primary() == nil {
		return nil, fmt.Errorf("endpoint: primary table missing")
	}

	primary, err := s.GetTable(e.Primary().FQID())
	if err != nil {
		return nil, fmt.Errorf("endpoint: %w", err)
	}

	var out models.Rows
	for i := range primary.Data {
		rec := primary.AddTransformations(primary.Data[i])

		// vertical data-minimalization.
		if rec.MatchPrimary(filter) {
			out = append(out, rec)
		}
	}

	if list := e.Joins; len(list) > 0 {
		out, err = joins.ProcessJoins(out, primary, list, filter, s)
		if err != nil {
			return nil, fmt.Errorf("endpoint: %w", err)
		}
	}

	if out == nil || len(out) == 0 {
		return nil, fmt.Errorf("endpoint: no matching records found for {%v}", filter)
	}

	// horizontal data-minimalization.
	return out.MatchFields(fields), nil
}
