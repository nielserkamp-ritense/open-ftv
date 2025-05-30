package joins

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/filters"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/memory/models"
)

func (j *Join) joinOnFields() (models.Rows, error) {
	out := make(models.Rows, 0, len(j.targetData))
	tableFilter := filters.NewTableFilter(j.source.Definition(), j.filter)
	filtered := len(tableFilter) > 0

	for i := range j.targetData {
		in := j.targetData[i]

		if key := in.KeyValueForFields(j.targetFields); key != "" {
			var joinData models.Rows
			for _, rec := range j.source.Data {
				if rec.KeyValueForFields(j.sourceFields) == key && rec.MatchFilter(tableFilter) {
					joinData = append(joinData, rec)
				}
			}

			if len(joinData) > 0 && len(j.def.Joins) > 0 {
				var err error
				if joinData, err = ProcessJoins(joinData, j.source, j.def.Joins, j.filter, j.meta); err != nil {
					return nil, err
				}
			}

			if result := j.merge(in, joinData, filtered); result != nil {
				out = append(out, result)
			}
		}
	}

	return out, nil
}
