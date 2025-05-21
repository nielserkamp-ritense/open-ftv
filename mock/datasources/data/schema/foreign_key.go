package schema

import (
	"strings"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/types"
)

type ForeignKey struct {
	Index
	ForeignTable string
	// hidden fields
	foreignTable  *Table
	foreignFields []*Field
}

// IterateForeignFields iterates over the foreign table field definitions and calls the closure for each field.
func (fk *ForeignKey) IterateForeignFields(f func(*Field)) {
	fk.Fix(nil, nil)
	for _, field := range fk.foreignFields {
		f(field)
	}
}

// MarshalJSON implements the JSON Marshaler interface.
func (fk *ForeignKey) MarshalJSON() ([]byte, error) {
	d2 := encodeFK{
		ID:           fk.ID,
		Description:  fk.Description,
		ForeignTable: fk.ForeignTable,
		Fields:       fk.Fields,
		Orders:       fk.Orders,
	}
	return json.Marshal(&d2)
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (fk *ForeignKey) UnmarshalJSON(b []byte) error {
	var d2 encodeFK
	if err := json.Unmarshal(b, &d2); err != nil {
		return err
	}

	fk.ID = d2.ID
	fk.Description = d2.Description
	fk.ForeignTable = d2.ForeignTable
	fk.Fields = d2.Fields
	fk.Orders = d2.Orders

	return nil
}

// MarshalYAML implements the YAML Marshaler interface.
func (fk *ForeignKey) MarshalYAML() ([]byte, error) {
	d2 := encodeFK{
		ID:           fk.ID,
		Description:  fk.Description,
		ForeignTable: fk.ForeignTable,
		Fields:       fk.Fields,
		Orders:       fk.Orders,
	}
	return yaml.Marshal(&d2)
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (fk *ForeignKey) UnmarshalYAML(b []byte) error {
	var d2 encodeFK
	if err := yaml.Unmarshal(b, &d2); err != nil {
		return err
	}

	fk.ID = d2.ID
	fk.Description = d2.Description
	fk.ForeignTable = d2.ForeignTable
	fk.Fields = d2.Fields
	fk.Orders = d2.Orders

	return nil
}

type encodeFK struct {
	ID           string            `json:"id" yaml:"id"`
	Description  string            `json:"description,omitempty" yaml:"description,omitempty"`
	ForeignTable string            `json:"foreignTable" yaml:"foreignTable"`
	Fields       []string          `json:"fields" yaml:"fields"`
	Orders       []types.OrderType `json:"orders,omitempty" yaml:"orders,omitempty"`
}

// Fix (re)sets the parent-child relationships for this object.
func (fk *ForeignKey) Fix(t *Table, tables map[string]*Table) {
	fk.mutex.Lock()
	fk.fix(t, tables)
	fk.mutex.Unlock()
}

func (fk *ForeignKey) fix(t *Table, tables map[string]*Table) {
	fk.Index.fix(t)

	if tables == nil && t.parentSource != nil {
		tables = t.parentSource.tables
	}

	if tables != nil {
		fk.foreignTable = tables[fk.ForeignTable]
	}

	if fk.foreignTable != nil && len(fk.foreignFields) == 0 {
		fk.foreignFields = make([]*Field, 0, len(fk.Fields))
		for _, id := range fk.Fields {
			parts := strings.Split(id, ":")

			var f *Field
			switch len(parts) {
			case 1:
				f = fk.foreignTable.fields[parts[0]]
			default:
				f = fk.foreignTable.fields[parts[1]]
			}

			if f != nil {
				fk.foreignFields = append(fk.foreignFields, f)
			}
		}
	}
}
