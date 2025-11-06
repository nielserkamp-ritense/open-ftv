package schema

import (
	"strings"
	"sync"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/compare"
)

// Index represents an index on a datasource table.
type Index struct {
	Parent
	Unique      bool
	Description string
	Fields      []string
	Orders      []enums.OrderType
	// hidden fields
	mutex       sync.Mutex
	parentTable *Table
	fields      []*Field
}

// IterateFields iterates over the field definitions in the index and calls the closure for each field.
func (i *Index) IterateFields(f func(*Field)) {
	i.Fix(nil)
	for _, field := range i.fields {
		f(field)
	}
}

// Equal returns true if this index list of fields matches the other list of fields.
//
// The order of the elements can be different as long as all elements exist in both lists.
//
// For foreign keys there may be a field mapping defined, so we need to determine
func (i *Index) Equal(other []string) bool {
	mine := make([]string, len(i.Fields))
	for j := range i.Fields {
		if k := strings.Index(i.Fields[j], ":"); k >= 0 {
			mine[j] = i.Fields[j][k+1:]
		} else {
			mine[j] = i.Fields[j]
		}
	}

	return compare.StringsEqual(mine, other)
}

// MarshalJSON implements the JSON Marshaler interface.
func (i *Index) MarshalJSON() ([]byte, error) {
	d2 := encodeIndex{
		ID:          i.ID,
		Unique:      i.Unique,
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
	i.Unique = d2.Unique
	i.Description = d2.Description
	i.Fields = d2.Fields
	i.Orders = d2.Orders

	return nil
}

// MarshalYAML implements the YAML Marshaler interface.
func (i *Index) MarshalYAML() ([]byte, error) {
	d2 := encodeIndex{
		ID:          i.ID,
		Unique:      i.Unique,
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
	i.Unique = d2.Unique
	i.Description = d2.Description
	i.Fields = d2.Fields
	i.Orders = d2.Orders

	return nil
}

type encodeIndex struct {
	ID          string            `json:"id" yaml:"id"`
	Unique      bool              `json:"unique,omitempty" yaml:"unique,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Fields      []string          `json:"fields,omitempty" yaml:"fields"`
	Orders      []enums.OrderType `json:"orders,omitempty" yaml:"orders,omitempty"`
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
			parts := strings.Split(id, ":")
			if len(parts) > 0 {
				i.fields = append(i.fields, t.fields[parts[0]])
			}
		}
	}
}
