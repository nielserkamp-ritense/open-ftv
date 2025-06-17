package joins

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/context"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/memory/models"
)

// ProcessJoins returns the given input data extended with the given list of joins,
// or an error if it fails to execute a join.
func ProcessJoins(in models.Rows, target *models.Table, list []*schema.Join, reqCtx *context.RequestContext, meta store.MetaReader) (models.Rows, error) {
	out := in

	var err error

	for _, j := range list {
		joinTarget := target
		if p := j.GetTarget(); p != nil {
			if joinTarget, err = meta.GetTable(p.FQID()); err != nil {
				return nil, fmt.Errorf("processJoins: target [%s] not found", j.Target)
			}
		}

		if len(j.Fields) > 0 {
			if out, err = New(joinTarget, out, j, meta).JoinOnFields(j.Fields, reqCtx.Filter, reqCtx.Params); err != nil {
				return nil, err
			}
		} else {
			def := j.GetSource()
			if def == nil {
				return nil, fmt.Errorf("processJoins: source definition [%s] not found", j.Source)
			}

			if fk := def.FindForeignKey(joinTarget.Definition()); fk != nil {
				if out, err = New(joinTarget, out, j, meta).JoinOnFK(fk, reqCtx.Filter, reqCtx.Params); err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("processJoins: invalid join definition [%s]: no matching foreign key", j.Source)
			}
		}
	}

	return out, nil
}
