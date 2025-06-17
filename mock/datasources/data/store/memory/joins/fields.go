package joins

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/context"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/memory/models"
)

func (j *Join) joinOnFields() (models.Rows, error) {
	out := make(models.Rows, 0, len(j.targetData))
	filtered := j.filter != nil

	for i := range j.targetData {
		in := j.targetData[i]

		if key := in.KeyValueForFields(j.targetFields); key != "" {
			var joinData models.Rows
			for _, rec := range j.source.Data {
				if rec.KeyValueForFields(j.sourceFields) == key {
					rec = j.source.AddTransformations(rec, j.params)
					if !filtered || rec.MatchJoin(j.def, j.filter) {
						joinData = append(joinData, rec)
					}
				}
			}

			if len(joinData) > 0 && len(j.def.Joins) > 0 {
				var err error
				if joinData, err = ProcessJoins(joinData, j.source, j.def.Joins, &context.RequestContext{Filter: j.filter, Params: j.params}, j.meta); err != nil {
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
