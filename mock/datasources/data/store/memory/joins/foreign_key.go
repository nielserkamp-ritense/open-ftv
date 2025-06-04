package joins

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/memory/models"
)

func (j *Join) joinOnFK(fk *schema.ForeignKey) (models.Rows, error) {
	out := make(models.Rows, 0, len(j.targetData))
	filtered := j.filter != nil

	for i := range j.targetData {
		foreign := j.source.ForeignKeys[fk.ID]
		if foreign == nil {
			return nil, fmt.Errorf("joinOnFK: foreign key [%s] not found", fk.ID)
		}

		in := j.targetData[i]
		if key := in.KeyValueForFK(fk); key != "" {
			recs := foreign[key]

			var joinData models.Rows
			if len(recs) > 0 {
				joinData = make(models.Rows, 0, len(recs))
				for _, rec := range recs {
					if !filtered || rec.MatchJoin(j.def, j.filter) {
						joinData = append(joinData, rec)
					}
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
