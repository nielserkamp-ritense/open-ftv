package joins

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/memory/models"
)

func (j *Join) merge(target *models.Row, source models.Rows, filtered bool) *models.Row {
	switch {
	case source != nil && len(source) > 0:
		return j.mergeOneOrMore(target, source) // we found one or more joins that passed the filter.
	case filtered:
		return j.mergeFiltered(target) // we found no joins that passed the filter.
	default:
		return j.mergeNotFound(target) // we found no joins.
	}
}

func (j *Join) mergeOneOrMore(target *models.Row, source models.Rows) *models.Row {
	if !j.def.IncludeJoinFields {
		source = source.RemoveFields(j.sourceFields)
	}

	switch j.def.Type {
	case enums.OptionalSibling, enums.ForcedSibling:
		if len(source) == 1 {
			// join a single sibling.
			return j.mergeSibling(target, source[0])
		}

		// TODO: join multiple siblings instead of forcing parent/child join.
		return j.mergeChildren(target, source)

	default:
		// join a slice of children.
		return j.mergeChildren(target, source)
	}
}

func (j *Join) mergeFiltered(target *models.Row) *models.Row {
	if j.def.Type == enums.OptionalParentChild || j.def.Type == enums.OptionalSibling {
		// keep the row as-is.
		return target
	}

	// join data was filtered out, and it's mandatory, the target must not be in the result set.
	// so we just skip adding it to the output!
	return nil
}

func (j *Join) mergeNotFound(target *models.Row) *models.Row {
	switch j.def.Type {
	case enums.ForcedParentChild:
		// force an empty slice of children.
		return j.mergeChildren(target, make(models.Rows, 0))

	case enums.ForcedSibling:
		// force a single empty sibling.
		rec := j.source.DummyRecord()
		if !j.def.IncludeJoinFields {
			rec = rec.RemoveFields(j.sourceFields)
		}
		return j.mergeSibling(target, rec)

	default:
		if j.def.QualifiedFields {
			// keep the row as-is, but with qualified field identifiers.
			return target.Qualified(j.target.Definition().ID)
		}

		// keep the row as-is.
		return target
	}
}

func (j *Join) mergeSibling(target, source *models.Row) *models.Row {
	if j.def.QualifiedFields {
		return target.JoinSiblingQualified(j.target.Definition().ID, j.source.Definition().ID, source)
	}
	return target.JoinSibling(source)
}

func (j *Join) mergeChildren(target *models.Row, source models.Rows) *models.Row {
	child := &models.Row{Data: map[string]any{j.def.GetJoinID(): source}}
	return target.JoinSibling(child)
}
