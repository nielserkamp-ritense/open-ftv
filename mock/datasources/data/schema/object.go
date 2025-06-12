package schema

import (
	"sync"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/maps"
)

// Object represents a datasource object; e.g., a datasource table or structured field.
type Object struct {
	Parent
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Fields      []*Field `json:"fields" yaml:"fields"`
	// hidden fields
	mutex       sync.Mutex
	parentTable *Table
	parentField *Field
	fields      map[string]*Field
}

// GetField returns the field definition for the given identifier.
func (o *Object) GetField(id string) *Field {
	return o.fields[id]
}

// IterateFields iterates over the field definitions in the object and calls the closure for each field.
func (o *Object) IterateFields(f func(*Field)) {
	o.Fix(nil, nil, nil)
	maps.ProcessOrdered(o.fields, func(_ string, v *Field) {
		f(v)
	})
}

// Fix (re)sets the parent-child relationships for this object.
func (o *Object) Fix(parent *Parent, table *Table, field *Field) {
	o.mutex.Lock()
	o.fix(parent, table, field)
	o.mutex.Unlock()
}

func (o *Object) fix(parent *Parent, table *Table, field *Field) {
	if parent != nil {
		o.parent = parent
	}
	if table != nil {
		o.parentTable = table
	}
	if field != nil {
		o.parentField = field
	}

	o.fields = make(map[string]*Field, len(o.Fields))
	for _, f2 := range o.Fields {
		f2.Fix(o, table, field)
		o.fields[f2.ID] = f2
	}
}
