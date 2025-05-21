package models

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/writer/csv"
)

// Rows is a convenience type for a slice of table data rows.
type Rows []*Row

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
