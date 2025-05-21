package memory

import (
	"fmt"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/memory/models"
)

// SetDataspace implements the Storage interface.
func (s *storage) SetDataspace(def *schema.Dataspace) {
	s.spaceDef = def
	s.space = models.NewSpace(def)

	for _, source := range def.DataSources {
		s.sourceDefs[strings.ToLower(source.ID)] = source
		s.sources[strings.ToLower(source.ID)] = models.NewSource(source)
	}
}

// AddDatasource implements the Storage interface.
func (s *storage) AddDatasource(def *schema.Datasource) {
	s.sourceDefs[strings.ToLower(def.ID)] = def
	s.sources[strings.ToLower(def.ID)] = models.NewSource(def)
}

// DatasourceExists implements the Storage interface.
func (s *storage) DatasourceExists(id string) bool {
	return s.GetDatasource(id) != nil
}

// TableExists implements the Storage interface.
func (s *storage) TableExists(id string) bool {
	t, err := s.GetTable(id)
	return err == nil && t != nil
}

// GetDataspace implements the Storage interface.
func (s *storage) GetDataspace() *models.Dataspace {
	return s.space
}

// GetDatasources implements the Storage interface.
func (s *storage) GetDatasources() map[string]*models.Datasource {
	return s.sources
}

// GetDatasource implements the Storage interface.
func (s *storage) GetDatasource(id string) *models.Datasource {
	return s.sources[strings.ToLower(id)]
}

// GetTable implements the Storage interface.
func (s *storage) GetTable(id string) (*models.Table, error) {
	parts := strings.Split(strings.ToLower(id), ".")

	var t *models.Table
	var err error

	switch len(parts) {
	case 1:
		t, err = s.findUnqualifiedTable(parts[0])
	case 2:
		t, err = s.findTable(parts[0], parts[1])
	case 3:
		if s.spaceDef != nil && strings.EqualFold(parts[0], s.spaceDef.ID) {
			t, err = s.findTable(parts[1], parts[2])
		}
	default:
		return nil, fmt.Errorf("invalid table id [%s]", id)
	}

	return t, err
}
