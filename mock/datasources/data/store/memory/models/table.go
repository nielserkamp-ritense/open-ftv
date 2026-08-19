package models

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/transforming"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// Table contains the data rows of a data table.
type Table struct {
	Data        Rows
	PK          map[string]*Row
	Indexes     map[string]map[string]Rows
	ForeignKeys map[string]map[string]Rows
	// hidden fields
	def           *schema.Table
	mutex         sync.RWMutex
	modifiedSince time.Time
}

// TableFromData returns a set of structured records from the given data using the given object definition.
//
// Errors are ignored, so prepare your input data well.
func TableFromData(def *schema.Table, data []map[string]any) *Table {
	t := newTable(def, len(data))
	for i := range data {
		_ = t.createRow(RowFromData(&def.Object, data[i]))
	}
	return t
}

// TableFromCSV returns a set of structured records from the given CSV data using the given object definition.
//
// Errors are ignored, so prepare your input file well.
func TableFromCSV(def *schema.Table, csv [][]string) *Table {
	t := newTable(def, len(csv)-1)
	if len(csv) > 1 {
		for i := range csv[1:] {
			_ = t.createRow(RowFromCSV(&def.Object, csv[0], csv[i+1]))
		}
	}
	return t
}

// Definition returns the table definition.
func (t *Table) Definition() *schema.Table {
	return t.def
}

// ModifiedSince returns the timestamp the table was last modified.
func (t *Table) ModifiedSince() time.Time {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.modifiedSince
}

// AsRow converts the table definition into an exportable row.
func (t *Table) AsRow() *Row {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	out := &Row{Data: make(map[string]any), def: tableRowDef}

	out.Data[fieldDefFQDN.ID] = t.def.FQDN()
	out.Data[fieldDefID.ID] = t.def.ID

	if t.def.Description != "" {
		out.Data[fieldDefDescription.ID] = t.def.Description
	}

	if l := len(t.def.Fields); l > 0 {
		out2 := make([]*Row, 0, l)
		for _, f2 := range t.def.Fields {
			out2 = append(out2, fieldAsRow(f2))
		}
		out.Data[fieldDefFields.ID] = out2
	}

	if t.def.PrimaryKey != nil {
		out.Data[fieldDefPK.ID] = indexAsRow(t.def.PrimaryKey)
	}

	if l := len(t.def.SecondaryIndexes); l > 0 {
		out2 := make(Rows, 0, l)
		for _, ix := range t.def.SecondaryIndexes {
			out2 = append(out2, indexAsRow(ix))
		}
		out.Data[fieldDefIndexes.ID] = out2
	}

	if l := len(t.def.ForeignKeys); l > 0 {
		out2 := make(Rows, 0, l)
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
func (t *Table) AddTransformations(in *Row, params map[string]any) *Row {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if len(t.def.Transforms) == 0 {
		return in
	}

	l := len(in.Data) + len(t.def.Transforms)
	out := &Row{
		Data:        make(map[string]any, l),
		qualifiers:  make(map[string]string, l),
		def:         in.def,
		isQualified: in.isQualified,
	}

	for k, v := range in.Data {
		out.Data[k] = v
		out.qualifiers[k] = in.qualifiers[k]
	}

	for _, transform := range t.def.Transforms {
		out.Data, out.qualifiers = transforming.Execute(out.Data, out.qualifiers, out.isQualified, transform, params)
	}
	return out
}

// DummyRecord returns an empty row according to the data table definition.
func (t *Table) DummyRecord() *Row {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	fields := t.def.Fields

	l := len(fields)
	out := &Row{
		Data:       make(map[string]any, l),
		qualifiers: make(map[string]string, l),
		def:        &t.def.Object,
	}

	for i := range fields {
		f := fields[i]
		out.Data[f.ID] = nil
		out.qualifiers[f.ID] = fmt.Sprintf("%s.%s", t.def.ID, f.ID)
	}
	return out
}

// CreateRow adds a new record to the data table.
func (t *Table) CreateRow(r *Row) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.createRow(r)
}

func (t *Table) createRow(r *Row) error {
	def := t.def

	var pk string

	ixKeys := make([]string, len(def.SecondaryIndexes))
	fkKeys := make([]string, len(def.ForeignKeys))

	if def.PrimaryKey != nil {
		pk = r.KeyValueForIndex(def.PrimaryKey)
		if _, ok := t.PK[pk]; ok {
			return fmt.Errorf("duplicate primary key: %s", pk)
		}
	}

	for i, ix := range def.SecondaryIndexes {
		key := r.KeyValueForIndex(ix)
		ixKeys[i] = key
		if ix.Unique {
			index := t.Indexes[ix.ID]
			if _, ok := index[key]; ok {
				return fmt.Errorf("duplicate secondary index: %s with primary key: %s", key, pk)
			}
		}
	}

	for i, fk := range def.ForeignKeys {
		fkKeys[i] = r.KeyValueForFK(fk, false)

		// TODO: check foreign key exists
	}

	t.Data = append(t.Data, r)

	if def.PrimaryKey != nil {
		t.PK[pk] = r
	}

	for i, ix := range def.SecondaryIndexes {
		key := ixKeys[i]
		index := t.Indexes[ix.ID]
		index[key] = append(index[key], r)
		t.Indexes[ix.ID] = index
	}

	for i, fk := range def.ForeignKeys {
		key := fkKeys[i]
		foreign := t.ForeignKeys[fk.ID]
		foreign[key] = append(foreign[key], r)
		t.ForeignKeys[fk.ID] = foreign
	}

	t.modifiedSince = time.Now().UTC()

	return nil
}

// DeleteRow removes a record from the data table.
func (t *Table) DeleteRow(pk []any) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.deleteRow(pk)
}

func (t *Table) deleteRow(pk []any) error {
	def := t.def
	if def.PrimaryKey == nil {
		return fmt.Errorf("missing primary key")
	}

	var old *Row

	key := KeyFromData(pk, t.Definition().PrimaryKey)

	var ok bool
	if old, ok = t.PK[key]; !ok {
		return fmt.Errorf("primary key %v not found", pk)
	}

	delete(t.PK, key)

	for i, rec := range t.Data {
		if rec == old {
			t.Data = append(t.Data[:i], t.Data[i+1:]...)
			break
		}
	}

	// TODO: remove record from secondary indexes
	// TODO: remove record from foreign keys

	t.modifiedSince = time.Now().UTC()

	return nil
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
	if cap < 2 {
		cap = 2
	}

	t := &Table{
		def:           def,
		modifiedSince: time.Now().UTC(),
		Data:          make(Rows, 0, cap),
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
