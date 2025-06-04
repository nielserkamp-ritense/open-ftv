package filtering

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/compare"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

// MatchOnPrimaryData returns true if the primary result data passes the filter operation.
func (f *FieldValueFilter) MatchOnPrimaryData(data map[string]any) bool {
	if f.Level == enums.JoinLevel {
		// not a primary test, so skip.
		return true
	}
	return f.compare(data)
}

// MatchOnJoinData returns true if the result data from a join passes the filter operation.
func (f *FieldValueFilter) MatchOnJoinData(join *schema.Join, data map[string]any) bool {
	switch {
	case f.Level == enums.PrimaryLevel:
		return true // primary level.
	case f.join != nil && join != f.join:
		return true // different join.
	case join.GetSource() != f.table:
		return true // different table (should never happen).
	}
	return f.compare(data)
}

func (f *FieldValueFilter) compare(data map[string]any) bool {
	return compare.Compare(compare.Params{
		Type:        f.field.Type,
		Compare:     f.Compare,
		Insensitive: f.Insensitive,
		Input:       data[f.field.ID],
		Value:       f.Value,
		Values:      f.Values,
		RX:          f.rx,
	})
}
