package schema

import (
	"strings"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
)

type ForeignKey struct {
	Index
	ForeignTable string
	SourceFields []string
	TargetFields []string
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
	Orders       []enums.OrderType `json:"orders,omitempty" yaml:"orders,omitempty"`
}

// Fix (re)sets the parent-child relationships for this object.
func (fk *ForeignKey) Fix(t *Table, tables map[string]*Table) {
	fk.mutex.Lock()
	fk.fix(t, tables)
	fk.mutex.Unlock()
}

func (fk *ForeignKey) fix(t *Table, tables map[string]*Table) {
	fk.Index.fix(t)

	fk.SourceFields, fk.TargetFields = SplitFields(fk.Fields)

	if tables == nil && t.parentSource != nil {
		tables = t.parentSource.tables
	}

	if tables != nil {
		fk.foreignTable = tables[fk.ForeignTable]
	}

	if fk.foreignTable != nil && len(fk.foreignFields) == 0 {
		fk.foreignFields = make([]*Field, 0, len(fk.Fields))
		for _, id := range fk.TargetFields {
			if f := fk.foreignTable.fields[id]; f != nil {
				fk.foreignFields = append(fk.foreignFields, f)
			}
		}
	}
}

// SplitFields splits foreign key fields into their source and target parts.
//
// The source is the field in the table the foreign key is defined for.
// The target is the field in the foreign table.
func SplitFields(list []string) ([]string, []string) {
	sources := make([]string, len(list))
	targets := make([]string, len(list))

	for i := range list {
		parts := strings.Split(list[i], ":")
		switch len(parts) {
		case 1:
			sources[i] = parts[0]
			targets[i] = parts[0]
		default:
			sources[i] = parts[0]
			targets[i] = parts[1]
		}
	}

	return sources, targets
}
