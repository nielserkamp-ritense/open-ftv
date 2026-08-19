package models

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"

// NewSpace instantiates a new data source.
func NewSpace(def *schema.Dataspace) *Dataspace {
	ds := &Dataspace{def: def, Sources: make(map[string]*Datasource)}
	if def != nil {
		for _, source := range def.DataSources {
			ds.Sources[source.ID] = NewSource(source)
		}
	}
	return ds
}

// Dataspace contains the data of a data space.
type Dataspace struct {
	Sources map[string]*Datasource
	// hidden fields
	def *schema.Dataspace
}

// AddDatasource adds a table to the data source from the given data.
func (s *Dataspace) AddDatasource(source *Datasource) {
	s.Sources[source.def.ID] = source
}

// AsRow converts the datasource definition into an exportable row.
func (s *Dataspace) AsRow() *Row {
	out := &Row{Data: make(map[string]any), def: dataspaceRowDef}

	out.Data[fieldDefFQDN.ID] = s.def.FQDN()
	out.Data[fieldDefID.ID] = s.def.ID

	if s.def.Description != "" {
		out.Data[fieldDefDescription.ID] = s.def.Description
	}

	if len(s.Sources) > 0 {
		out2 := make(Rows, 0, len(s.Sources))
		for _, source := range s.Sources {
			out2 = append(out2, source.AsRow())
		}
		out.Data[fieldDefSources.ID] = out2
	}

	return out
}
