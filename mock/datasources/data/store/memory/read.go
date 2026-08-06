package memory

import (
	"fmt"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/context"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/matching"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store/memory/joins"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store/memory/models"
)

// SelectPK implements the Reader interface.
func (s *storage) SelectPK(tableID string, pk []any, matcher matching.FieldMatcher) (*models.Row, *time.Time, error) {
	table, err := s.GetTable(tableID)
	if err != nil {
		return nil, nil, fmt.Errorf("selectPK: %w", err)
	}

	if len(table.Definition().PrimaryKey.Fields) != len(pk) {
		return nil, nil, fmt.Errorf("selectPK: field count mismatch")
	}

	modifiedSince := table.ModifiedSince()

	key := models.KeyFromData(pk, table.Definition().PrimaryKey)
	if rec := table.PK[key]; rec != nil {
		return rec.MatchFields(matcher), &modifiedSince, nil
	}
	return nil, nil, fmt.Errorf("selectPK: primary key %v not found", pk)
}

// SelectIX implements the Reader interface.
func (s *storage) SelectIX(tableID string, id string, keys []any, matcher matching.FieldMatcher) (models.Rows, *time.Time, error) {
	table, err := s.GetTable(tableID)
	if err != nil {
		return nil, nil, fmt.Errorf("selectIX: %w", err)
	}

	index := table.Definition().SecondaryIndex(id)
	if index == nil {
		return nil, nil, fmt.Errorf("selectIX: index [%s] not found", id)
	}

	if len(index.Fields) != len(keys) {
		return nil, nil, fmt.Errorf("selectIX: field count mismatch")
	}

	modifiedSince := table.ModifiedSince()

	key := models.KeyFromData(keys, index)
	if indexData := table.Indexes[id]; indexData != nil {
		if list := indexData[key]; len(list) > 0 {
			return list.MatchFields(matcher), &modifiedSince, nil
		}
	}
	return nil, nil, fmt.Errorf("selectIX: keys %v in index [%s] not found", keys, id)
}

// Search implements the Reader interface.
func (s *storage) Search(tableID string, reqCtx *context.RequestContext) (models.Rows, *time.Time, error) {
	table, err := s.GetTable(tableID)
	if err != nil {
		return nil, nil, fmt.Errorf("search: %w", err)
	}

	modifiedSince := table.ModifiedSince()

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
		return out, &modifiedSince, nil
	}
	return nil, nil, fmt.Errorf("search: no matching records found for {%v}", reqCtx.Filter)
}

// GetEndpoint implements the Reader interface.
func (s *storage) GetEndpoint(e *schema.Endpoint, ctx *context.RequestContext) (models.Rows, *time.Time, error) {
	if e.Primary() == nil {
		return nil, nil, fmt.Errorf("endpoint: primary table missing")
	}

	primary, err := s.GetTable(e.Primary().FQID())
	if err != nil {
		return nil, nil, fmt.Errorf("endpoint: %w", err)
	}

	modifiedSince := primary.ModifiedSince()

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
			return nil, nil, fmt.Errorf("endpoint: %w", err)
		}
		modifiedSince = s.determineModifiedSince(modifiedSince, list)
	}

	if out == nil || len(out) == 0 {
		return models.Rows{}, nil, nil
	}

	// horizontal data-minimalization.
	return out.MatchFields(ctx.Matcher), &modifiedSince, nil
}

// GetEndpointByPK implements the Reader interface.
//
// It returns a single record identified by its primary key, running it through the same table
// transformations, joins and field matching that GetEndpoint applies, so both GET paths stay
// consistent (e.g. a table with masking/pseudonymisation transforms behaves the same either way).
func (s *storage) GetEndpointByPK(e *schema.Endpoint, pk []any, ctx *context.RequestContext) (*models.Row, *time.Time, error) {
	if e.Primary() == nil {
		return nil, nil, fmt.Errorf("endpoint: primary table missing")
	}

	primary, err := s.GetTable(e.Primary().FQID())
	if err != nil {
		return nil, nil, fmt.Errorf("endpoint: %w", err)
	}

	rec, modifiedSince, err := s.SelectPK(e.Primary().FQID(), pk, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("endpoint: %w", err)
	}

	out := models.Rows{primary.AddTransformations(rec, ctx.Params)}

	if list := e.Joins; len(list) > 0 {
		if out, err = joins.ProcessJoins(out, primary, list, ctx, s); err != nil {
			return nil, nil, fmt.Errorf("endpoint: %w", err)
		}

		ms := s.determineModifiedSince(*modifiedSince, list)
		modifiedSince = &ms
	}

	out = out.MatchFields(ctx.Matcher)
	if len(out) == 0 {
		return nil, nil, fmt.Errorf("endpoint: primary key %v not found", pk)
	}

	return out[0], modifiedSince, nil
}

func (s *storage) determineModifiedSince(in time.Time, list []*schema.Join) time.Time {
	out := in

	for _, j := range list {
		if table, err := s.GetTable(j.Target); err == nil {
			if t := table.ModifiedSince(); t.After(in) {
				out = t
			}
		}

		out = s.determineModifiedSince(out, j.Joins)
	}

	return out
}
