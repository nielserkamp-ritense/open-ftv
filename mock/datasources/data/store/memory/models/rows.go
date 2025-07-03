package models

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/matching"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/writer/csv"
)

// Rows is a convenience type for a list of table data rows.
type Rows []*Row

// MatchFields returns a deep copy of the table data with only those fields that pass the given field matcher.
func (r Rows) MatchFields(fields matching.FieldMatcher) Rows {
	if fields == nil {
		return r
	}

	out := make(Rows, len(r))
	for i, row := range r {
		out[i] = row.MatchFields(fields)
	}
	return out
}

// RemoveFields returns a deep copy of the table data with the given fields removed from each row.
func (r Rows) RemoveFields(fields []string) Rows {
	l := len(fields)
	if l == 0 {
		return r
	}

	m := make(map[string]struct{}, l)
	for i := range fields {
		m[fields[i]] = struct{}{}
	}

	out := make(Rows, len(r))
	for i, row := range r {
		out[i] = row.RemoveFieldMap(m)
	}
	return out
}

// MarshalCSV implements the CSV marshaler interface.
func (r Rows) MarshalCSV() ([]byte, error) {
	if len(r) == 0 {
		return []byte{}, nil
	}

	enc := csv.NewBytesEncoder()
	writeKeys(r[0], enc)
	for i := range r {
		writeRecord(r[i], enc)
	}
	return enc.Bytes(), nil
}

// Encode implements the CSV Encoder interface.
func (r Rows) Encode() any {
	out := make([]any, len(r))
	for i := range r {
		out[i] = r[i].Encode()
	}
	return out
}
