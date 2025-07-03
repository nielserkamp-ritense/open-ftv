// Package memory contains functionality to store and retrieve data-space and/or -source definitions and the actual data in memory.
package memory

import (
	"fmt"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store/memory/models"
)

// New instantiates a new memory store.
func New(ds *schema.Dataspace) store.Storage {
	s := &storage{
		spaceDef:   ds,
		sourceDefs: make(map[string]*schema.Datasource),
		sources:    make(map[string]*models.Datasource),
		tables:     make(map[string]*models.Table),
		endpoints:  make(map[string]*schema.Endpoint),
	}

	if ds != nil {
		ds.Fix()

		for _, source := range ds.DataSources {
			s.sourceDefs[strings.ToLower(source.ID)] = source
			s.sources[strings.ToLower(source.ID)] = models.NewSource(source)
		}
	}

	return s
}

func (s *storage) findTableDef(sourceID, tableID string) (*schema.Table, error) {
	source := s.sourceDefs[strings.ToLower(sourceID)]
	if source == nil {
		return nil, fmt.Errorf("source id [%s] unknown", sourceID)
	}

	out := source.Table(tableID)
	if out == nil {
		return nil, fmt.Errorf("table id [%s] unknown", tableID)
	}
	return out, nil
}

func (s *storage) findTable(sourceID, tableID string) (*models.Table, error) {
	source := s.sources[strings.ToLower(sourceID)]
	if source == nil {
		return nil, fmt.Errorf("source id [%s] unknown", sourceID)
	}

	out := source.Tables[strings.ToLower(tableID)]
	if out == nil {
		return nil, fmt.Errorf("table id [%s] unknown", tableID)
	}
	return out, nil
}

func (s *storage) findUnqualifiedTable(tableID string) (*models.Table, error) {
	var out *models.Table

	for _, source := range s.sources {
		if table := source.Tables[strings.ToLower(tableID)]; table != nil {
			if out != nil {
				return nil, fmt.Errorf("table id [%s] not unique", tableID)
			}
			out = table
		}
	}

	if out == nil {
		return nil, fmt.Errorf("table id [%s] unknown", tableID)
	}
	return out, nil
}

type storage struct {
	spaceDef   *schema.Dataspace
	sourceDefs map[string]*schema.Datasource
	space      *models.Dataspace
	sources    map[string]*models.Datasource
	tables     map[string]*models.Table
	endpoints  map[string]*schema.Endpoint
}
