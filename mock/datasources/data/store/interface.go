package store

import (
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/context"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/matching"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store/memory/models"
)

// Storage represents the interface for storing and retrieving data-space/-source definitions as well as manipulating data tables.
type Storage interface {
	MetaLoader
	MetaReader

	Loader
	Maintainer
	Reader
}

// MetaLoader represents the interface for setting the dataspace and adding datasource definitions.
type MetaLoader interface {
	SetDataspace(def *schema.Dataspace)
	AddDatasource(def *schema.Datasource)
	AddEndpoint(def *schema.Endpoint)
}

// MetaReader represents the interface for searching and retrieving data-space/-source definitions.
type MetaReader interface {
	IterateEndpoints(f func(def *schema.Endpoint))
	DatasourceExists(id string) bool
	TableExists(id string) bool

	GetDataspace() *models.Dataspace
	GetDatasources() map[string]*models.Datasource
	GetDatasource(id string) *models.Datasource
	GetTable(id string) (*models.Table, error)
}

// Loader represents the interface for storing data-space/-source definitions as well as the actual data tables.
type Loader interface {
	AddTableFromData(sourceID, tableID string, data []map[string]any) error
	AddTableFromCSV(sourceID, tableID string, csv [][]string) error
}

// Maintainer represents the interface for creating, updating and deleting records in/from data tables.
type Maintainer interface {
	CreateRecord(tableID string, record *models.Row) error
	UpdateRecord(tableID string, pk []any, record *models.Row) error
	DeleteRecord(tableID string, pk []any) error
}

// Reader represents the interface for searching and retrieving records from data tables.
type Reader interface {
	SelectPK(tableID string, pk []any, matcher matching.FieldMatcher) (*models.Row, *time.Time, error)
	SelectIX(tableID string, id string, keys []any, matcher matching.FieldMatcher) (models.Rows, *time.Time, error)
	Search(tableID string, ctx *context.RequestContext) (models.Rows, *time.Time, error)
	GetEndpoint(e *schema.Endpoint, ctx *context.RequestContext) (models.Rows, *time.Time, error)
}
