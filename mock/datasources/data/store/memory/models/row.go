package models

import (
	"regexp"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/types"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/writer/csv"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
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
	def *schema.Object
}

// FieldString returns the value of a data element as a string.
func (r *Row) FieldString(id string) string {
	if v, ok := r.Data[id]; ok {
		return convert.AnyToString(v)
	}
	return ""
}

// MatchFilter returns true if the record matches the given filter.
func (r *Row) MatchFilter(filter map[string]any) bool {
	for k, v := range filter {
		s1 := r.FieldString(k)

		var ok bool
		switch t := v.(type) {
		case *regexp.Regexp:
			ok = t.MatchString(s1)
		default:
			ok = s1 == convert.AnyToString(v)
		}

		if !ok {
			return false
		}
	}

	return true
}

// MatchFields returns the record with only those fields that pass the given field matcher.
func (r *Row) MatchFields(matcher types.FieldMatcher) *Row {
	if matcher.Always() {
		return r
	}

	out := &Row{Data: make(map[string]any), def: &schema.Object{Fields: make([]*schema.Field, 0, len(r.def.Fields))}}
	r.def.IterateFields(func(f *schema.Field) {
		if matcher.Match(f.ID) {
			out.Data[f.ID] = r.Data[f.ID]
			out.def.Fields = append(out.def.Fields, f)
		}
	})
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
