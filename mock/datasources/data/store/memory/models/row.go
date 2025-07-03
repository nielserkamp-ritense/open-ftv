package models

import (
	"bytes"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/filtering"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/matching"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/writer/csv"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// RowFromData returns a structured record from the given data using the given object definition.
func RowFromData(def *schema.Object, data map[string]any) *Row {
	r := &Row{Data: make(map[string]any, len(data)), def: def}

	def.IterateFields(func(f *schema.Field) {
		if d, ok := data[f.ID]; ok {
			if len(f.Fields) > 0 {
				if m2, ok2 := d.(map[string]any); ok2 {
					r.Data[f.ID] = RowFromData(&f.Object, m2)
				} else {
					// TODO: error!
				}
			} else {
				r.Data[f.ID] = f.ConvertValue(d)
			}
		}
	})

	return r
}

// RowFromCSV returns a structured record from the given CSV data using the given object definition.
func RowFromCSV(def *schema.Object, header, data []string) *Row {
	r := &Row{Data: make(map[string]any, len(data)), def: def}

	keys := make(map[string]int, len(header))
	for i := range header {
		keys[header[i]] = i
	}

	def.IterateFields(func(f *schema.Field) {
		if i, ok := keys[f.ID]; ok {
			if len(f.Fields) > 0 {
				// TODO: error or unmarshal from JSON?
			} else {
				if s := data[i]; s != "" {
					r.Data[f.ID] = f.ConvertValue(s)
				}
			}
		}
	})

	return r
}

// Row contains the data of a single row from a data table.
type Row struct {
	Data map[string]any
	// hidden fields
	def         *schema.Object
	isQualified bool
}

// FieldString returns the value of a data element as a string.
func (r *Row) FieldString(id string) string {
	if v, ok := r.Data[id]; ok {
		return convert.AnyToString(v)
	}
	return ""
}

// KeyValueForIndex returns a concatenated key of all the field values for the given index.
func (r *Row) KeyValueForIndex(index *schema.Index) string {
	b := bytes.Buffer{}

	index.IterateFields(func(field *schema.Field) {
		if r.isQualified {
			b.WriteString(r.FieldString(field.FQID()))
		} else {
			b.WriteString(r.FieldString(field.ID))
		}
		b.WriteByte('|')
	})

	if b.Len() > 0 {
		b.Truncate(b.Len() - 1)
	}
	return b.String()
}

// KeyValueForFK returns a concatenated key of all the field values for the given foreign key.
func (r *Row) KeyValueForFK(fk *schema.ForeignKey) string {
	b := bytes.Buffer{}

	fk.IterateFields(func(field *schema.Field) {
		if r.isQualified {
			b.WriteString(r.FieldString(field.FQID()))
		} else {
			b.WriteString(r.FieldString(field.ID))
		}
		b.WriteByte('|')
	})

	if b.Len() > 0 {
		b.Truncate(b.Len() - 1)
	}
	return b.String()
}

// KeyValueForFields returns a concatenated key of all the field values for the given list of field keys.
func (r *Row) KeyValueForFields(keys []string) string {
	b := bytes.Buffer{}

	for i := range keys {
		b.WriteString(r.FieldString(keys[i]))
		b.WriteByte('|')
	}

	if b.Len() > 0 {
		b.Truncate(b.Len() - 1)
	}
	return b.String()
}

// MatchPrimary returns true if the record matches the given filter.
func (r *Row) MatchPrimary(filter filtering.Filterer) bool {
	return filter.MatchOnPrimaryData(r.Data)
}

// MatchJoin returns true if the record matches the given filter.
func (r *Row) MatchJoin(join *schema.Join, filter filtering.Filterer) bool {
	return filter.MatchOnJoinData(join, r.Data)
}

// MatchFields returns the record with only those fields that pass the given field matcher.
func (r *Row) MatchFields(matcher matching.FieldMatcher) *Row {
	if matcher == nil || matcher.Always() {
		return r
	}

	out := &Row{
		Data:        make(map[string]any),
		def:         r.def,
		isQualified: r.isQualified,
	}

	r.def.IterateFields(func(f *schema.Field) {
		if out.isQualified {
			if s := f.FQID(); matcher.Match(s) {
				out.Data[s] = r.Data[s]
			}
		} else {
			if s := f.ID; matcher.Match(s) {
				out.Data[s] = r.Data[s]
			}
		}
	})

	return out
}

// JoinSibling returns a deep copy of the row extended with the data from the joined row.
func (r *Row) JoinSibling(r2 *Row) *Row {
	out := &Row{
		Data:        make(map[string]any, len(r.Data)+len(r2.Data)),
		def:         r.def,
		isQualified: r.isQualified,
	}

	for k, v := range r.Data {
		out.Data[k] = v
	}
	for k, v := range r2.Data {
		out.Data[k] = v
	}
	return out
}

// JoinSiblingQualified returns a deep copy of the row extended with the data from the joined row.
//
// Unlike JoinSibling, this function makes sure all field identifiers are fully qualified.
func (r *Row) JoinSiblingQualified(t1, t2 string, r2 *Row) *Row {
	out := r.Qualified(t1)

	if r2.isQualified {
		for k, v := range r2.Data {
			out.Data[k] = v
		}
	} else {
		for k, v := range r2.Data {
			out.Data[fmt.Sprintf("%s.%s", t2, k)] = v
		}
	}

	return out
}

// Qualified makes a deep copy of the record while adding the given table qualifier to the field identifiers (if needed).
func (r *Row) Qualified(tableID string) *Row {
	out := &Row{
		Data:        make(map[string]any, len(tableID)),
		def:         r.def,
		isQualified: true,
	}

	if r.isQualified {
		for k, v := range r.Data {
			out.Data[k] = v
		}
	} else {
		for k, v := range r.Data {
			out.Data[fmt.Sprintf("%s.%s", tableID, k)] = v
		}
	}

	return out
}

// RemoveFields returns a deep copy of the row with the given fields removed from the data.
func (r *Row) RemoveFields(fields []string) *Row {
	m := make(map[string]struct{}, len(fields))
	for i := range fields {
		m[fields[i]] = struct{}{}
	}
	return r.RemoveFieldMap(m)
}

// RemoveFieldMap returns a deep copy of the row with the given fields removed from the data.
func (r *Row) RemoveFieldMap(fields map[string]struct{}) *Row {
	out := &Row{
		Data:        make(map[string]any, len(fields)),
		def:         r.def,
		isQualified: r.isQualified,
	}

	for k, v := range r.Data {
		if _, ok := fields[k]; !ok {
			out.Data[k] = v
		}
	}
	return out
}

// MarshalJSON implements the JSON Marshaler interface.
func (r *Row) MarshalJSON() ([]byte, error) {
	return json.Marshal(removeNil(r.Data))
}

// MarshalYAML implements the YAML Marshaler interface.
func (r *Row) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(removeNil(r.Data))
}

// MarshalCSV implements the CSV marshaler interface.
func (r *Row) MarshalCSV() ([]byte, error) {
	enc := csv.NewBytesEncoder()
	writeKeys(r, enc)
	writeRecord(r, enc)
	return enc.Bytes(), nil
}

// Encode implements the CSV encoder interface.
func (r *Row) Encode() any {
	return r.Data
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (r *Row) UnmarshalJSON(b []byte) error {
	var data map[string]any
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	r.Data = removeNil(data)
	return nil
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (r *Row) UnmarshalYAML(b []byte) error {
	var data map[string]any
	if err := yaml.Unmarshal(b, &data); err != nil {
		return err
	}

	r.Data = removeNil(data)
	return nil
}

func removeNil(data map[string]any) map[string]any {
	out := make(map[string]any, len(data))
	for k, v := range data {
		if v != nil {
			out[k] = v
		}
	}
	return out
}
