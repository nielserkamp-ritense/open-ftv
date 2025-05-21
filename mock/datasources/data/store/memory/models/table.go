package models

import (
	"bytes"
	"regexp"
	"strings"
	"sync"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// Table contains the data rows of a data table.
type Table struct {
	Data    []*Row                     `json:"data" yaml:"data"`
	PK      map[string]*Row            `json:"primaryKey,omitempty" yaml:"primaryKey,omitempty"`
	Indexes map[string]map[string]Rows `json:"secondaryIndexes,omitempty" yaml:"secondaryIndexes,omitempty"`
	// hidden fields
	def   *schema.Table
	mutex sync.RWMutex
}

// TableFromData returns a set of structured records from the given data using the given object definition.
func TableFromData(def *schema.Table, data []map[string]any) *Table {
	t := newTable(def, len(data))
	for i := range data {
		t.createRecord(RowFromData(&def.Object, data[i]))
	}
	return t
}

// TableFromCSV returns a set of structured records from the given CSV data using the given object definition.
func TableFromCSV(def *schema.Table, csv [][]string) *Table {
	t := newTable(def, len(csv)-1)
	if len(csv) > 1 {
		for i := range csv[1:] {
			t.createRecord(RowFromCSV(&def.Object, csv[0], csv[i+1]))
		}
	}
	return t
}

// Definition returns the table definition.
func (t *Table) Definition() *schema.Table {
	return t.def
}

// MatchFilter returns true if the record matches the given filter.
func (t *Table) MatchFilter(filter map[string]any) bool {
	match := func(s string, v any) bool {
		switch tp := v.(type) {
		case *regexp.Regexp:
			return tp.MatchString(s)
		default:
			return s == convert.AnyToString(v)
		}
	}

	for k, v := range filter {
		switch strings.ToLower(k) {
		case "id":
			if !match(t.def.ID, v) {
				return false
			}
		case "description":
			if !match(t.def.Description, v) {
				return false
			}
		default:
			return false
		}
	}

	return true
}

// AsRecord converts the table definition into an exportable record.
func (t *Table) AsRecord() *Row {
	def := &schema.Object{Fields: []*schema.Field{
		&fieldDefFQDN,
		&fieldDefID,
		&fieldDefDescription,
		&fieldDefPK,
		&fieldDefIndexes,
		&fieldDefFK,
	}}

	out := &Row{Data: make(map[string]any), def: def}

	out.Data[fieldDefFQDN.ID] = t.def.FQDN()
	out.Data[fieldDefID.ID] = t.def.ID

	if t.def.Description != "" {
		out.Data[fieldDefDescription.ID] = t.def.Description
	}

	if t.def.PrimaryKey != nil {
		out.Data[fieldDefPK.ID] = indexAsRecord(t.def.PrimaryKey)
	}

	if len(t.def.SecondaryIndexes) > 0 {
		out2 := make(Rows, 0, len(t.def.SecondaryIndexes))
		for _, ix := range t.def.SecondaryIndexes {
			out2 = append(out2, indexAsRecord(ix))
		}
		out.Data[fieldDefIndexes.ID] = out2
	}

	if len(t.def.ForeignKeys) > 0 {
		out2 := make(Rows, 0, len(t.def.ForeignKeys))
		for _, fk := range t.def.ForeignKeys {
			out2 = append(out2, fkAsRecord(fk))
		}
		out.Data[fieldDefFK.ID] = out2
	}

	return out
}

// CreateRecord adds a new record to the data table.
func (t *Table) CreateRecord(r *Row) {
	t.mutex.Lock()
	t.createRecord(r)
	t.mutex.Unlock()
}

func (t *Table) createRecord(r *Row) {
	t.Data = append(t.Data, r)
	def := t.def

	if def.PrimaryKey != nil {
		key := KeyFromRecord(r, def.PrimaryKey)
		t.PK[key] = r
	}

	for _, ix := range def.SecondaryIndexes {
		key := KeyFromRecord(r, ix)
		index := t.Indexes[ix.ID]
		index[key] = append(index[key], r)
		t.Indexes[ix.ID] = index
	}

	// TODO: something with foreign keys?

}

// KeyFromRecord returns a concatenated key of all field values for the given index.
func KeyFromRecord(r *Row, index *schema.Index) string {
	b := bytes.Buffer{}

	index.IterateFields(func(field *schema.Field) {
		b.WriteString(r.FieldString(field.ID))
		b.WriteByte('|')
	})

	return removeLastChar(b.String())
}

// KeyFromData returns a concatenated key from the given values for the given index.
func KeyFromData(keys []any, index *schema.Index) string {
	b := bytes.Buffer{}

	index.IterateFields(func(field *schema.Field) {
		if len(keys) > 0 {
			b.WriteString(convert.AnyToString(keys[0]))
			keys = keys[1:]
		}
		b.WriteByte('|')
	})

	return removeLastChar(b.String())
}

func newTable(def *schema.Table, cap int) *Table {
	_ = def.Field("") // force an internal fix().

	if cap < 2 {
		cap = 2
	}

	t := &Table{
		def:  def,
		Data: make([]*Row, 0, cap),
	}

	if def.PrimaryKey != nil {
		t.PK = make(map[string]*Row, cap)
	}

	if l := len(def.SecondaryIndexes); l > 0 {
		t.Indexes = make(map[string]map[string]Rows, l)

		for _, ix := range def.SecondaryIndexes {
			t.Indexes[ix.ID] = make(map[string]Rows)
		}
	}

	// TODO: something with foreign keys?

	return t
}

func removeLastChar(in string) string {
	if l := len(in); l > 0 {
		return in[:l-1]
	}
	return in
}
