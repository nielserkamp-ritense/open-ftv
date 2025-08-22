package models

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

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
	l := len(data)
	r := &Row{
		Data:       make(map[string]any, l),
		qualifiers: make(map[string]string, l),
		def:        def,
	}

	def.IterateFields(func(f *schema.Field) {
		r.qualifiers[f.ID] = fmt.Sprintf("%s.%s", def.ID, f.ID)

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
	l := len(data)
	r := &Row{
		Data:       make(map[string]any, l),
		qualifiers: make(map[string]string, l),
		def:        def,
	}

	keys := make(map[string]int, len(header))
	for i := range header {
		keys[header[i]] = i
	}

	def.IterateFields(func(f *schema.Field) {
		r.qualifiers[f.ID] = fmt.Sprintf("%s.%s", def.ID, f.ID)

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

// NewJoin returns a row with a single field initialized for use as a joined column.
func NewJoin(id string, source Rows) *Row {
	return &Row{
		Data:       map[string]any{id: source},
		qualifiers: map[string]string{id: "@join." + id},
	}
}

// Row contains the data of a single row from a data table.
type Row struct {
	Data map[string]any
	// hidden fields
	qualifiers  map[string]string
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
func (r *Row) KeyValueForFK(fk *schema.ForeignKey, target bool) string {
	var fields []string
	if target {
		fields = slices.Clone(fk.TargetFields)
	} else {
		fields = slices.Clone(fk.SourceFields)
	}

	if r.isQualified {
		for i := range fields {
			fields[i] = fmt.Sprintf("%s.%s", fk.ForeignTable, fields[i])
		}
	}

	b := bytes.Buffer{}

	for i := range fields {
		b.WriteString(r.FieldString(fields[i]))
		b.WriteByte('|')
	}

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
		qualifiers:  make(map[string]string),
		def:         r.def,
		isQualified: r.isQualified,
	}

	for k, v := range r.Data {
		fqid := r.qualifiers[k]

		add := func(v2 any) {
			out.Data[k] = v2
			out.qualifiers[k] = fqid
		}

		if strings.HasPrefix(fqid, "@join.") {
			if r2, ok := v.(Rows); ok {
				if r2 = r2.MatchFields(matcher); len(r2) > 0 && len(r2[0].Data) > 0 {
					add(r2)
				}
			} else {
				if matcher.Match(k) {
					add(v)
				}
			}
		} else {
			if matcher.Match(k) || matcher.Match(fqid) {
				add(v)
			}
		}
	}

	return out
}

// JoinSibling returns a deep copy of the row extended with the data from the joined row.
func (r *Row) JoinSibling(r2 *Row) *Row {
	l := len(r.Data) + len(r2.Data)
	out := &Row{
		Data:        make(map[string]any, l),
		qualifiers:  make(map[string]string, l),
		def:         r.def,
		isQualified: r.isQualified,
	}

	for k, v := range r.Data {
		out.Data[k] = v
		out.qualifiers[k] = r.qualifiers[k]
	}
	for k, v := range r2.Data {
		out.Data[k] = v
		out.qualifiers[k] = r2.qualifiers[k]
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
			out.qualifiers[k] = r2.qualifiers[k]
		}
	} else {
		for k, v := range r2.Data {
			fqid := fmt.Sprintf("%s.%s", t2, k)
			out.Data[fqid] = v
			out.qualifiers[fqid] = fqid
		}
	}

	return out
}

// Qualified makes a deep copy of the record while adding the given table qualifier to the field identifiers (if needed).
func (r *Row) Qualified(tableID string) *Row {
	l := len(tableID)
	out := &Row{
		Data:        make(map[string]any, l),
		qualifiers:  make(map[string]string, l),
		def:         r.def,
		isQualified: true,
	}

	if r.isQualified {
		for k, v := range r.Data {
			out.Data[k] = v
			out.qualifiers[k] = r.qualifiers[k]
		}
	} else {
		for k, v := range r.Data {
			fqid := fmt.Sprintf("%s.%s", tableID, k)
			out.Data[fqid] = v
			out.qualifiers[fqid] = fqid
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
	l := len(fields)
	out := &Row{
		Data:        make(map[string]any, l),
		qualifiers:  make(map[string]string, l),
		def:         r.def,
		isQualified: r.isQualified,
	}

	for k, v := range r.Data {
		if _, ok := fields[k]; !ok {
			out.Data[k] = v
			out.qualifiers[k] = r.qualifiers[k]
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
