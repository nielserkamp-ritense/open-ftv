// Package schema contains the definitions for one or more data spaces and/or sources.
package schema

import (
	"sync"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
)

// Dataspace represents the details of a dataspace.
type Dataspace struct {
	Parent
	Description string
	DataSources []*Datasource
	// hidden fields
	mutex   sync.Mutex
	sources map[string]*Datasource
}

// Source returns the source definition for the given id.
func (d *Dataspace) Source(sourceID string) *Datasource {
	d.Fix()
	return d.sources[sourceID]
}

// MarshalJSON implements the JSON Marshaler interface.
func (d *Dataspace) MarshalJSON() ([]byte, error) {
	d2 := encodeDataspace{
		ID:          d.ID,
		Description: d.Description,
		DataSources: d.DataSources,
	}
	return json.Marshal(&d2)
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (d *Dataspace) UnmarshalJSON(b []byte) error {
	var d2 encodeDataspace
	if err := json.Unmarshal(b, &d2); err != nil {
		return err
	}

	d.ID = d2.ID
	d.Description = d2.Description
	d.DataSources = d2.DataSources

	for _, ds := range d2.DataSources {
		ds.fix(d)
	}

	return nil
}

// MarshalYAML implements the YAML Marshaler interface.
func (d *Dataspace) MarshalYAML() ([]byte, error) {
	d2 := encodeDataspace{
		ID:          d.ID,
		Description: d.Description,
		DataSources: d.DataSources,
	}
	return yaml.Marshal(&d2)
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (d *Dataspace) UnmarshalYAML(b []byte) error {
	var d2 encodeDataspace
	if err := yaml.Unmarshal(b, &d2); err != nil {
		return err
	}

	d.ID = d2.ID
	d.Description = d2.Description
	d.DataSources = d2.DataSources

	for _, ds := range d2.DataSources {
		ds.fix(d)
	}

	return nil
}

type encodeDataspace struct {
	ID          string        `json:"id" yaml:"id"`
	Description string        `json:"description,omitempty" yaml:"description,omitempty"`
	DataSources []*Datasource `json:"dataSources,omitempty" yaml:"dataSources,omitempty"`
}

// Fix (re)sets the parent-child relationships for this object.
func (d *Dataspace) Fix() {
	d.mutex.Lock()
	d.fix()
	d.mutex.Unlock()
}

func (d *Dataspace) fix() {
	d.sources = make(map[string]*Datasource, len(d.DataSources))
	for _, ds := range d.DataSources {
		d.sources[ds.ID] = ds
	}

	// fix the data sources after we have the full map!
	for _, ds := range d.sources {
		ds.Fix(d)
	}
}
