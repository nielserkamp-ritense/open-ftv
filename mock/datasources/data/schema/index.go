package schema

import (
	"sync"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/types"
)

// Index represents an index on a datasource table.
type Index struct {
	Parent
	Description string
	Fields      []string
	Orders      []types.OrderType
	// hidden fields
	mutex       sync.Mutex
	parentTable *Table
	fields      []*Field
}

// IterateFields iterates over the field definitions in the index and calls the closure for each field.
func (i *Index) IterateFields(f func(*Field)) {
	for _, field := range i.fields {
		f(field)
	}
}

// MarshalJSON implements the JSON Marshaler interface.
func (i *Index) MarshalJSON() ([]byte, error) {
	d2 := encodeIndex{
		ID:          i.ID,
		Description: i.Description,
		Fields:      i.Fields,
		Orders:      i.Orders,
	}
	return json.Marshal(&d2)
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (i *Index) UnmarshalJSON(b []byte) error {
	var d2 encodeIndex
	if err := json.Unmarshal(b, &d2); err != nil {
		return err
	}

	i.ID = d2.ID
	i.Description = d2.Description
	i.Fields = d2.Fields
	i.Orders = d2.Orders

	return nil
}

// MarshalYAML implements the YAML Marshaler interface.
func (i *Index) MarshalYAML() ([]byte, error) {
	d2 := encodeIndex{
		ID:          i.ID,
		Description: i.Description,
		Fields:      i.Fields,
		Orders:      i.Orders,
	}
	return yaml.Marshal(&d2)
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (i *Index) UnmarshalYAML(b []byte) error {
	var d2 encodeIndex
	if err := yaml.Unmarshal(b, &d2); err != nil {
		return err
	}

	i.ID = d2.ID
	i.Description = d2.Description
	i.Fields = d2.Fields
	i.Orders = d2.Orders

	return nil
}

type encodeIndex struct {
	ID          string            `json:"id" yaml:"id"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Fields      []string          `json:"fields,omitempty" yaml:"fields"`
	Orders      []types.OrderType `json:"orders,omitempty" yaml:"orders,omitempty"`
}

// Fix (re)sets the parent-child relationships for this object.
func (i *Index) Fix(t *Table) {
	i.mutex.Lock()
	i.fix(t)
	i.mutex.Unlock()
}

func (i *Index) fix(t *Table) {
	if t != nil {
		i.parent = &t.Parent
		i.parentTable = t

		i.fields = make([]*Field, 0, len(i.Fields))
		for _, id := range i.Fields {
			i.fields = append(i.fields, t.fields[id])
		}
	}
}
