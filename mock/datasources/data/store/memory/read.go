package memory

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/context"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/matching"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store/memory/joins"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store/memory/models"
)

// SelectPK implements the Reader interface.
func (s *storage) SelectPK(tableID string, pk []any, matcher matching.FieldMatcher) (*models.Row, error) {
	table, err := s.GetTable(tableID)
	if err != nil {
		return nil, fmt.Errorf("selectPK: %w", err)
	}

	if len(table.Definition().PrimaryKey.Fields) != len(pk) {
		return nil, fmt.Errorf("selectPK: field count mismatch")
	}

	key := models.KeyFromData(pk, table.Definition().PrimaryKey)
	if rec := table.PK[key]; rec != nil {
		return rec.MatchFields(matcher), nil
	}
	return nil, fmt.Errorf("selectPK: primary key %v not found", pk)
}

// SelectIX implements the Reader interface.
func (s *storage) SelectIX(tableID string, id string, keys []any, matcher matching.FieldMatcher) (models.Rows, error) {
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

	key := models.KeyFromData(keys, index)
	if indexData := table.Indexes[id]; indexData != nil {
		if list := indexData[key]; len(list) > 0 {
			return list.MatchFields(matcher), nil
		}
	}
	return nil, fmt.Errorf("selectIX: keys %v in index [%s] not found", keys, id)
}

// Search implements the Reader interface.
func (s *storage) Search(tableID string, reqCtx *context.RequestContext) (models.Rows, error) {
	table, err := s.GetTable(tableID)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	var out models.Rows
	for i := range table.Data {
		rec := table.Data[i]
		// vertical data-minimalization.
		if rec.MatchPrimary(reqCtx.Filter) {
			// horizontal data-minimalization.
			out = append(out, table.AddTransformations(rec, reqCtx.Params).MatchFields(reqCtx.Matcher))
		}
	}

	if len(out) > 0 {
		return out, nil
	}
	return nil, fmt.Errorf("search: no matching records found for {%v}", reqCtx.Filter)
}

// GetEndpoint implements the Reader interface.
func (s *storage) GetEndpoint(e *schema.Endpoint, ctx *context.RequestContext) (models.Rows, error) {
	if e.Primary() == nil {
		return nil, fmt.Errorf("endpoint: primary table missing")
	}

	primary, err := s.GetTable(e.Primary().FQID())
	if err != nil {
		return nil, fmt.Errorf("endpoint: %w", err)
	}

	var out models.Rows
	for i := range primary.Data {
		rec := primary.AddTransformations(primary.Data[i], ctx.Params)

		// vertical data-minimalization.
		if rec.MatchPrimary(ctx.Filter) {
			out = append(out, rec)
		}
	}

	if list := e.Joins; len(list) > 0 {
		out, err = joins.ProcessJoins(out, primary, list, ctx, s)
		if err != nil {
			return nil, fmt.Errorf("endpoint: %w", err)
		}
	}

	if out == nil || len(out) == 0 {
		return nil, fmt.Errorf("endpoint: no matching records found for {%v}", ctx.Filter)
	}

	// horizontal data-minimalization.
	return out.MatchFields(ctx.Matcher), nil
}
