package schema

import (
	"fmt"
	"strings"
	"sync"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"golang.org/x/exp/maps"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/matching"
)

// Endpoint represents the details of a custom API endpoint.
//
// The table list determines which tables from the given datasource are targeted.
// If multiple tables are specified, field identifiers should be qualified with the table id: "t1.f1".
//
// The field list can be used to determine which fields will be accessed.
// If the list is empty, it means all fields will be available.
// If a field (with or without wildcard characters) is prefixed with an exclamation character,
// that field or fields will be excluded.
// If a field contains wildcard characters ('?','*'),
// all fields that match the wildcard expression will be included
// or excluded if the expression starts with an exclamation mark.
// To just exclude one or more fields, use []string{"*", "!f1", "!f2"}.
// An excluded field takes precedence over an included field.
type Endpoint struct {
	Version     uint8            // major version of the endpoint (e.g. v1, v2, ...).
	Type        enums.MethodType // the type of call to handle.
	CalledAs    enums.MethodType // the method the endpoint will be called with.
	Path        string           // the path (excluding the version) for the endpoint.
	FullVersion string           // the full version of the endpoint (e.g. 1.0.0, 1.2.1, ...).
	Description string           // description of the endpoint.
	Datasource  string           // the datasource the endpoint will access.
	Table       string           // the primary table within the datasource the endpoint will access.
	Joins       []*Join          // optional joins tables the endpoint will access.
	Fields      []string         // the fields within the tables the endpoint will access.
	Filter      map[string]any   // optional fixed filter for the endpoint.
	// hidden fields
	mutex         sync.Mutex
	datasource    *Datasource       // the datasource of the tables.
	primary       *Table            // primary table in a join.
	includeFields map[string]*Field // list of fields to include.
	excludeFields map[string]*Field // list of fields to exclude.
}

// UID returns a unique identifier for the endpoint based on the major version and the path.
func (e *Endpoint) UID() string {
	return fmt.Sprintf("/v%d/%s", e.Version, strings.Trim(e.Path, "/"))
}

// GetDatasource returns the datasource for this endpoint.
func (e *Endpoint) GetDatasource() *Datasource {
	return e.datasource
}

// Primary returns the primary table for this endpoint.
func (e *Endpoint) Primary() *Table {
	return e.primary
}

// FieldIncluded returns true if the given fully qualified field should be included.
func (e *Endpoint) FieldIncluded(fqdn string) bool {
	if _, ok := e.excludeFields[fqdn]; ok {
		return false
	}
	if len(e.includeFields) > 0 {
		_, ok := e.includeFields[fqdn]
		return ok
	}
	return true
}

// MarshalJSON implements the JSON Marshaler interface.
func (e *Endpoint) MarshalJSON() ([]byte, error) {
	e2 := encodeEndpoint{
		Version:     e.Version,
		Type:        e.Type,
		CalledAs:    e.CalledAs,
		Path:        e.Path,
		FullVersion: e.FullVersion,
		Description: e.Description,
		Datasource:  e.Datasource,
		Table:       e.Table,
		Joins:       e.Joins,
		Fields:      e.Fields,
		Filter:      e.Filter,
	}
	return json.Marshal(&e2)
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (e *Endpoint) UnmarshalJSON(b []byte) error {
	var e2 encodeEndpoint
	if err := json.Unmarshal(b, &e2); err != nil {
		return err
	}

	e.Version = e2.Version
	e.Type = e2.Type
	e.CalledAs = e2.CalledAs
	e.Path = e2.Path
	e.FullVersion = e2.FullVersion
	e.Description = e2.Description
	e.Datasource = e2.Datasource
	e.Table = e2.Table
	e.Joins = e2.Joins
	e.Fields = e2.Fields
	e.Filter = e2.Filter

	return nil
}

// MarshalYAML implements the YAML Marshaler interface.
func (e *Endpoint) MarshalYAML() ([]byte, error) {
	e2 := encodeEndpoint{
		Version:     e.Version,
		Type:        e.Type,
		CalledAs:    e.CalledAs,
		Path:        e.Path,
		FullVersion: e.FullVersion,
		Description: e.Description,
		Datasource:  e.Datasource,
		Table:       e.Table,
		Joins:       e.Joins,
		Fields:      e.Fields,
		Filter:      e.Filter,
	}
	return yaml.Marshal(&e2)
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (e *Endpoint) UnmarshalYAML(b []byte) error {
	var e2 encodeEndpoint
	if err := yaml.Unmarshal(b, &e2); err != nil {
		return err
	}

	e.Version = e2.Version
	e.Type = e2.Type
	e.CalledAs = e2.CalledAs
	e.Path = e2.Path
	e.FullVersion = e2.FullVersion
	e.Description = e2.Description
	e.Datasource = e2.Datasource
	e.Table = e2.Table
	e.Joins = e2.Joins
	e.Fields = e2.Fields
	e.Filter = e2.Filter

	return nil
}

type encodeEndpoint struct {
	Version     uint8            `json:"version" yaml:"version"`
	Type        enums.MethodType `json:"type" yaml:"type"`
	CalledAs    enums.MethodType `json:"calledAs" yaml:"calledAs"`
	Path        string           `json:"path" yaml:"path"`
	FullVersion string           `json:"fullVersion,omitempty" yaml:"fullVersion,omitempty"`
	Description string           `json:"description,omitempty" yaml:"description,omitempty"`
	Datasource  string           `json:"datasource,omitempty" yaml:"datasource,omitempty"`
	Table       string           `json:"table,omitempty" yaml:"table,omitempty"`
	Joins       []*Join          `json:"joins,omitempty" yaml:"joins,omitempty"`
	Fields      []string         `json:"fields,omitempty" yaml:"fields,omitempty"`
	Filter      map[string]any   `json:"filter,omitempty" yaml:"filter,omitempty"`
}

// Fix (re)sets the parent-child relationships for this object.
func (e *Endpoint) Fix(ds *Datasource) {
	e.mutex.Lock()
	e.fix(ds)
	e.mutex.Unlock()
}

func (e *Endpoint) fix(ds *Datasource) {
	if e.includeFields == nil {
		e.includeFields = make(map[string]*Field)
	}
	if e.excludeFields == nil {
		e.excludeFields = make(map[string]*Field)
	}

	if e.Type == 0 {
		e.Type = enums.GetMethod
	}
	if e.CalledAs == 0 {
		e.CalledAs = e.Type
	}

	if ds != nil {
		e.datasource = ds
		e.primary = nil

		maps.Clear(e.includeFields)
		maps.Clear(e.excludeFields)

		if t := ds.Table(e.Table); t != nil {
			e.primary = t
		}

		for _, j := range e.Joins {
			j.fix(ds)
		}

		// fix fields after the table and joins!
		for _, id := range e.Fields {
			e.fixField(id)
		}

		// make sure the field filter gave a result, and if not, we include all fields!
		if len(e.includeFields) == 0 && len(e.excludeFields) == 0 {
			for _, t := range e.datasource.Tables {
				for _, field := range t.Fields {
					e.includeFields[field.FQDN()] = field
				}
			}
		}
	}
}

func (e *Endpoint) fixField(id string) {
	var exclude bool
	if strings.HasPrefix(id, "!") {
		exclude = true
		id = id[1:]
	}

	parts := strings.Split(id, ".")

	testTable := func(t *Table, m matching.FieldMatcher) {
		for _, field := range t.Fields {
			if m.Match(field.ID) {
				if exclude {
					e.excludeFields[field.FQDN()] = field
				} else {
					e.includeFields[field.FQDN()] = field
				}
			}
		}
	}

	switch len(parts) {
	case 1:
		m1 := matching.NewFieldMatcher(parts[0])
		testTable(e.primary, m1)
		for _, j := range e.Joins {
			testTable(j.source, m1)
		}

	case 2:
		m1 := matching.NewFieldMatcher(parts[0])
		m2 := matching.NewFieldMatcher(parts[1])
		for _, t := range e.datasource.Tables {
			if m1.Match(t.ID) {
				testTable(t, m2)
			}
		}
	}
}
