package schema

import (
	"sync"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
)

type ForeignKey struct {
	Parent
	Description  string
	Fields       []string
	ForeignTable string
	// hidden fields
	mutex        sync.Mutex
	parentTable  *Table
	fields       []*Field
	foreignTable *Table
}

// MarshalJSON implements the JSON Marshaler interface.
func (f *ForeignKey) MarshalJSON() ([]byte, error) {
	d2 := encodeFK{
		ID:           f.ID,
		Description:  f.Description,
		Fields:       f.Fields,
		ForeignTable: f.ForeignTable,
	}
	return json.Marshal(&d2)
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (f *ForeignKey) UnmarshalJSON(b []byte) error {
	var d2 encodeFK
	if err := json.Unmarshal(b, &d2); err != nil {
		return err
	}

	f.ID = d2.ID
	f.Description = d2.Description
	f.Fields = d2.Fields
	f.ForeignTable = d2.ForeignTable

	return nil
}

// MarshalYAML implements the YAML Marshaler interface.
func (f *ForeignKey) MarshalYAML() ([]byte, error) {
	d2 := encodeFK{
		ID:           f.ID,
		Description:  f.Description,
		Fields:       f.Fields,
		ForeignTable: f.ForeignTable,
	}
	return yaml.Marshal(&d2)
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (f *ForeignKey) UnmarshalYAML(b []byte) error {
	var d2 encodeFK
	if err := yaml.Unmarshal(b, &d2); err != nil {
		return err
	}

	f.ID = d2.ID
	f.Description = d2.Description
	f.Fields = d2.Fields
	f.ForeignTable = d2.ForeignTable

	return nil
}

type encodeFK struct {
	ID           string   `json:"id" yaml:"id"`
	Description  string   `json:"description,omitempty" yaml:"description,omitempty"`
	Fields       []string `json:"fields" yaml:"fields"`
	ForeignTable string   `json:"foreignTable" yaml:"foreignTable"`
}

// Fix (re)sets the parent-child relationships for this object.
func (f *ForeignKey) Fix(t *Table, tables map[string]*Table) {
	f.mutex.Lock()
	f.fix(t, tables)
	f.mutex.Unlock()
}

func (f *ForeignKey) fix(t *Table, tables map[string]*Table) {
	if t != nil {
		f.parent = &t.Parent
		f.parentTable = t

		f.fields = make([]*Field, 0, len(f.Fields))
		for _, id := range f.Fields {
			f.fields = append(f.fields, t.fields[id])
		}
	}

	if tables != nil {
		f.foreignTable = tables[f.ForeignTable]
	}
}
