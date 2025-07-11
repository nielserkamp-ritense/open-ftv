package joins

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/filtering"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store/memory/models"
)

// New instantiates a new join operation.
func New(target *models.Table, targetData models.Rows, joinDef *schema.Join, meta store.MetaReader) *Join {
	return &Join{
		target:     target,
		targetData: targetData,
		def:        joinDef,
		meta:       meta,
	}
}

// JoinOnFK executes the join operation based on the given foreign key and optional filter.
func (j *Join) JoinOnFK(fk *schema.ForeignKey, filter filtering.Filterer, params map[string]any) (models.Rows, error) {
	var err error
	if j.source, err = j.meta.GetTable(j.def.Source); err != nil || j.source == nil {
		if j.source == nil {
			return nil, fmt.Errorf("joinOnFK: source [%s] not found", j.def.Source)
		}
		return nil, err
	}

	j.filter = filter
	j.sourceFields, j.targetFields = fk.SourceFields, fk.TargetFields
	j.params = params

	return j.joinOnFK(fk)
}

// JoinOnFields executes the join operation based on the given list of fields and optional filter.
func (j *Join) JoinOnFields(fields []string, filter filtering.Filterer, params map[string]any) (models.Rows, error) {
	var err error
	if j.source, err = j.meta.GetTable(j.def.Source); err != nil || j.source == nil {
		if j.source == nil {
			return nil, fmt.Errorf("joinOnFields: source [%s] not found", j.def.Source)
		}
		return nil, err
	}

	j.filter = filter
	j.sourceFields, j.targetFields = schema.SplitFields(fields)
	j.params = params

	return j.joinOnFields()
}

// Join contains the parameters for a join operation.
type Join struct {
	target       *models.Table
	targetData   models.Rows
	source       *models.Table
	def          *schema.Join
	filter       filtering.Filterer
	params       map[string]any
	sourceFields []string
	targetFields []string
	meta         store.MetaReader
}
