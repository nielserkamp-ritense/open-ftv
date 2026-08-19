package schema

import (
	"strings"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
)

// Datasource represents the configuration of a datasource.
type Datasource struct {
	Parent
	Description string
	Tables      []*Table
	// hidden fields
	ds     *Dataspace
	tables map[string]*Table
}

// Table returns the table definition for the given id.
func (d *Datasource) Table(tableID string) *Table {
	return d.tables[strings.ToLower(tableID)]
}

// MarshalJSON implements the JSON Marshaler interface.
func (d *Datasource) MarshalJSON() ([]byte, error) {
	d2 := encodeDatasource{
		ID:          d.ID,
		Description: d.Description,
		Tables:      d.Tables,
	}
	return json.Marshal(&d2)
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (d *Datasource) UnmarshalJSON(b []byte) error {
	var d2 encodeDatasource
	if err := json.Unmarshal(b, &d2); err != nil {
		return err
	}

	d.ID = d2.ID
	d.Description = d2.Description
	d.Tables = d2.Tables

	d.Fix(nil)

	return nil
}

// MarshalYAML implements the YAML Marshaler interface.
func (d *Datasource) MarshalYAML() ([]byte, error) {
	d2 := encodeDatasource{
		ID:          d.ID,
		Description: d.Description,
		Tables:      d.Tables,
	}
	return yaml.Marshal(&d2)
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (d *Datasource) UnmarshalYAML(b []byte) error {
	var d2 encodeDatasource
	if err := yaml.Unmarshal(b, &d2); err != nil {
		return err
	}

	d.ID = d2.ID
	d.Description = d2.Description
	d.Tables = d2.Tables

	d.Fix(nil)

	return nil
}

type encodeDatasource struct {
	ID          string   `json:"id" yaml:"id"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Tables      []*Table `json:"tables,omitempty" yaml:"tables,omitempty"`
}

// Fix (re)sets the parent-child relationships for this object.
func (d *Datasource) Fix(ds *Dataspace) {
	if ds != nil {
		d.parent = &ds.Parent
		d.ds = ds
	}

	d.tables = make(map[string]*Table, len(d.Tables))
	for _, t := range d.Tables {
		d.tables[strings.ToLower(t.ID)] = t
	}

	// fix the tables *after* we have the full map!
	for _, t := range d.Tables {
		t.fixSelf(d)
	}

	// fix the foreign keys *after* every table has resolved its own fields,
	// as a foreign key points at the fields of another table!
	for _, t := range d.Tables {
		t.fixRelations(d)
	}
}
