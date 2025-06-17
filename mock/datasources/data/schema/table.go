package schema

import (
	"fmt"
	"strings"
	"sync"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
)

// Table represents a datasource table.
type Table struct {
	Object
	PrimaryKey       *Index
	SecondaryIndexes []*Index
	ForeignKeys      []*ForeignKey
	Transforms       []*Transformation
	// hidden fields
	mutex        sync.Mutex
	parentSource *Datasource
	indexes      map[string]*Index
	foreignKeys  map[string]*ForeignKey
	transforms   map[string]*Transformation
}

// FQID returns the fully qualified ID of the table.
func (t *Table) FQID() string {
	if t.parentSource != nil {
		return fmt.Sprintf("%s.%s", t.parentSource.ID, t.ID)
	}
	return t.ID
}

// Datasource returns the data source definition for the table.
func (t *Table) Datasource() *Datasource {
	return t.parentSource
}

// Field returns the field definition for the given id.
func (t *Table) Field(fieldID string) *Field {
	t.Fix(nil)
	return t.fields[fieldID]
}

// SecondaryIndex returns the secondary index definition for the given id.
func (t *Table) SecondaryIndex(id string) *Index {
	t.Fix(nil)
	return t.indexes[id]
}

// ForeignKey returns the foreign key definition for the given id.
func (t *Table) ForeignKey(id string) *ForeignKey {
	t.Fix(nil)
	return t.foreignKeys[id]
}

// Transformation returns the transformation definition for the given id.
func (t *Table) Transformation(id string) *Transformation {
	t.Fix(nil)
	return t.transforms[id]
}

// FindForeignKey returns the foreign key definition that matches the given index.
func (t *Table) FindForeignKey(foreign *Table) *ForeignKey {
	t.Fix(nil)

	for _, fk := range t.foreignKeys {
		if strings.EqualFold(foreign.ID, fk.ForeignTable) && foreign.PrimaryKey.Equal(fk.Fields) {
			return fk
		}
	}
	return nil
}

// MarshalJSON implements the JSON Marshaler interface.
func (t *Table) MarshalJSON() ([]byte, error) {
	t2 := encodeTable{
		ID:               t.ID,
		Description:      t.Description,
		Fields:           t.Fields,
		PrimaryKey:       t.PrimaryKey,
		SecondaryIndexes: t.SecondaryIndexes,
		ForeignKeys:      t.ForeignKeys,
		Transforms:       t.Transforms,
	}
	return json.Marshal(&t2)
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (t *Table) UnmarshalJSON(b []byte) error {
	var t2 encodeTable
	if err := json.Unmarshal(b, &t2); err != nil {
		return err
	}

	t.ID = t2.ID
	t.Description = t2.Description
	t.Fields = t2.Fields
	t.PrimaryKey = t2.PrimaryKey
	t.SecondaryIndexes = t2.SecondaryIndexes
	t.ForeignKeys = t2.ForeignKeys
	t.Transforms = t2.Transforms

	return nil
}

// MarshalYAML implements the YAML Marshaler interface.
func (t *Table) MarshalYAML() ([]byte, error) {
	t2 := encodeTable{
		ID:               t.ID,
		Description:      t.Description,
		Fields:           t.Fields,
		PrimaryKey:       t.PrimaryKey,
		SecondaryIndexes: t.SecondaryIndexes,
		ForeignKeys:      t.ForeignKeys,
		Transforms:       t.Transforms,
	}
	return yaml.Marshal(&t2)
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (t *Table) UnmarshalYAML(b []byte) error {
	var t2 encodeTable
	if err := yaml.Unmarshal(b, &t2); err != nil {
		return err
	}

	t.ID = t2.ID
	t.Description = t2.Description
	t.Fields = t2.Fields
	t.PrimaryKey = t2.PrimaryKey
	t.SecondaryIndexes = t2.SecondaryIndexes
	t.ForeignKeys = t2.ForeignKeys
	t.Transforms = t2.Transforms

	return nil
}

type encodeTable struct {
	ID               string            `json:"id" yaml:"id"`
	Description      string            `json:"description,omitempty" yaml:"description,omitempty"`
	Fields           []*Field          `json:"fields,omitempty" yaml:"fields,omitempty"`
	PrimaryKey       *Index            `json:"primaryKey,omitempty" yaml:"primaryKey,omitempty"`
	SecondaryIndexes []*Index          `json:"secondaryIndexes,omitempty" yaml:"secondaryIndexes,omitempty"`
	ForeignKeys      []*ForeignKey     `json:"foreignKeys,omitempty" yaml:"foreignKeys,omitempty"`
	Transforms       []*Transformation `json:"transformations,omitempty" yaml:"transformations,omitempty"`
}

// Fix (re)sets the parent-child relationships for this object.
func (t *Table) Fix(d *Datasource) {
	t.mutex.Lock()
	t.fix(d)
	t.mutex.Unlock()
}

func (t *Table) fix(d *Datasource) {
	if d == nil {
		t.Object.fix(nil, nil, nil)
	} else {
		t.Object.fix(&d.Parent, nil, nil)
		t.parentSource = d
	}

	if t.PrimaryKey != nil {
		t.PrimaryKey.Fix(t)
	}

	// fix indexes after we have the full map of fields!
	t.indexes = make(map[string]*Index, len(t.SecondaryIndexes))
	for _, index := range t.SecondaryIndexes {
		index.Fix(t)
		t.indexes[index.ID] = index
	}

	if d != nil {
		// fix foreign keys after we have the full map of fields!
		t.foreignKeys = make(map[string]*ForeignKey, len(t.ForeignKeys))
		for _, fk := range t.ForeignKeys {
			fk.Fix(t, d.tables)
			t.foreignKeys[fk.ID] = fk
		}
	}

	t.transforms = make(map[string]*Transformation, len(t.Transforms))
	for _, transform := range t.Transforms {
		transform.Fix(t)
		t.transforms[transform.ID] = transform
	}
}
