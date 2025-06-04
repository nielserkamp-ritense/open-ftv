package models

import (
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

// NewSource instantiates a new datasource.
func NewSource(def *schema.Datasource) *Datasource {
	ds := &Datasource{def: def, Tables: make(map[string]*Table)}
	for _, table := range def.Tables {
		ds.Tables[strings.ToLower(table.ID)] = newTable(table, 0)
	}
	return ds
}

// Datasource contains all data for a datasource.
type Datasource struct {
	Tables map[string]*Table
	// hidden fields
	def *schema.Datasource
}

// AddTableFromData adds a table to the datasource from the given data.
func (s *Datasource) AddTableFromData(def *schema.Table, data []map[string]any) {
	s.Tables[strings.ToLower(def.ID)] = TableFromData(def, data)
}

// AddTableFromCSV adds a table to the datasource from the given CSV data.
func (s *Datasource) AddTableFromCSV(def *schema.Table, csv [][]string) {
	s.Tables[strings.ToLower(def.ID)] = TableFromCSV(def, csv)
}

// Definition returns the definition of the datasource.
func (s *Datasource) Definition() *schema.Datasource {
	return s.def
}

// AsRow converts the datasource definition into an exportable row.
func (s *Datasource) AsRow() *Row {
	def := &schema.Object{Fields: []*schema.Field{
		&fieldDefFQDN,
		&fieldDefID,
		&fieldDefDescription,
		&fieldDefTables,
	}}

	out := &Row{Data: make(map[string]any), def: def}

	out.Data[fieldDefFQDN.ID] = s.def.FQDN()
	out.Data[fieldDefID.ID] = s.def.ID

	if s.def.Description != "" {
		out.Data[fieldDefDescription.ID] = s.def.Description
	}

	if len(s.Tables) > 0 {
		out2 := make(Rows, len(s.Tables))
		for _, t := range s.Tables {
			out2 = append(out2, t.AsRow())
		}
		out.Data[fieldDefTables.ID] = out2
	}

	return out
}
