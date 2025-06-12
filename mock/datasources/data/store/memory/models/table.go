package models

import (
	"bytes"
	"sync"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/transforming"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// Table contains the data rows of a data table.
type Table struct {
	Data        Rows
	PK          map[string]*Row
	Indexes     map[string]map[string]Rows
	ForeignKeys map[string]map[string]Rows
	// hidden fields
	def   *schema.Table
	mutex sync.RWMutex
}

// TableFromData returns a set of structured records from the given data using the given object definition.
func TableFromData(def *schema.Table, data []map[string]any) *Table {
	t := newTable(def, len(data))
	for i := range data {
		t.createRow(RowFromData(&def.Object, data[i]))
	}
	return t
}

// TableFromCSV returns a set of structured records from the given CSV data using the given object definition.
func TableFromCSV(def *schema.Table, csv [][]string) *Table {
	t := newTable(def, len(csv)-1)
	if len(csv) > 1 {
		for i := range csv[1:] {
			t.createRow(RowFromCSV(&def.Object, csv[0], csv[i+1]))
		}
	}
	return t
}

// Definition returns the table definition.
func (t *Table) Definition() *schema.Table {
	return t.def
}

// AsRow converts the table definition into an exportable row.
func (t *Table) AsRow() *Row {
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
		out.Data[fieldDefPK.ID] = indexAsRow(t.def.PrimaryKey)
	}

	if len(t.def.SecondaryIndexes) > 0 {
		out2 := make(Rows, 0, len(t.def.SecondaryIndexes))
		for _, ix := range t.def.SecondaryIndexes {
			out2 = append(out2, indexAsRow(ix))
		}
		out.Data[fieldDefIndexes.ID] = out2
	}

	if len(t.def.ForeignKeys) > 0 {
		out2 := make(Rows, 0, len(t.def.ForeignKeys))
		for _, fk := range t.def.ForeignKeys {
			out2 = append(out2, fkAsRow(fk))
		}
		out.Data[fieldDefFK.ID] = out2
	}

	return out
}

// AddTransformations returns a deep copy of the given table data adding any transformations defined for the table.
//
// If there are no transformations defined for the table, the input table data is returned as-is.
func (t *Table) AddTransformations(in *Row) *Row {
	if len(t.def.Transforms) == 0 {
		return in
	}

	out := &Row{
		Data:        make(map[string]any, len(in.Data)+len(t.def.Transforms)),
		def:         in.def,
		isQualified: in.isQualified,
	}

	for k, v := range in.Data {
		out.Data[k] = v
	}

	for _, transform := range t.def.Transforms {
		out.Data = transforming.Execute(out.Data, out.isQualified, transform)
	}
	return out
}

// DummyRecord returns an empty row according to the data table definition.
func (t *Table) DummyRecord() *Row {
	fields := t.def.Fields
	out := &Row{Data: make(map[string]any, len(fields)), def: &t.def.Object}
	for i := range fields {
		out.Data[fields[i].ID] = nil
	}
	return out
}

// CreateRow adds a new record to the data table.
func (t *Table) CreateRow(r *Row) {
	t.mutex.Lock()
	t.createRow(r)
	t.mutex.Unlock()
}

func (t *Table) createRow(r *Row) {
	t.Data = append(t.Data, r)
	def := t.def

	if def.PrimaryKey != nil {
		key := r.KeyValueForIndex(def.PrimaryKey)
		t.PK[key] = r
	}

	for _, ix := range def.SecondaryIndexes {
		key := r.KeyValueForIndex(ix)
		index := t.Indexes[ix.ID]
		index[key] = append(index[key], r)
		t.Indexes[ix.ID] = index
	}

	for _, fk := range def.ForeignKeys {
		key := r.KeyValueForFK(fk)
		foreign := t.ForeignKeys[fk.ID]
		foreign[key] = append(foreign[key], r)
		t.ForeignKeys[fk.ID] = foreign
	}
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

	if b.Len() > 0 {
		b.Truncate(b.Len() - 1)
	}
	return b.String()
}

func newTable(def *schema.Table, cap int) *Table {
	def.Fix(nil)

	if cap < 2 {
		cap = 2
	}

	t := &Table{
		def:  def,
		Data: make(Rows, 0, cap),
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

	if l := len(def.ForeignKeys); l > 0 {
		t.ForeignKeys = make(map[string]map[string]Rows, l)

		for _, ix := range def.ForeignKeys {
			t.ForeignKeys[ix.ID] = make(map[string]Rows)
		}
	}

	return t
}
