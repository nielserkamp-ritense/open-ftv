package network

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
)

func (r *runner) decodeRelation(obj *RelationObject) {
	if data := findElement(splitKeys(obj.Base), r.data); data != nil {
		r.decodeRelationData(data, obj)
	}
}

func (r *runner) decodeRelationData(data any, obj *RelationObject) {
	switch t := data.(type) {
	case []any:
		r.decodeRelationsSlice(t, obj)
	case map[string]any:
		r.decodeRelationMap(t, obj)
	default:
		r.processRelation(
			obj.SubjectTypeValue, obj.SubjectIdValue,
			obj.PredicateTypeValue, obj.PredicateIdValue,
			obj.ObjectTypeValue, obj.ObjectIdValue,
			obj)
	}
}

func (r *runner) decodeRelationsSlice(m []any, obj *RelationObject) {
	for i := range m {
		r.decodeRelationData(m[i], obj)
	}
}

func (r *runner) decodeRelationMap(m map[string]any, obj *RelationObject) {
	sType := codeOrValueString(obj.SubjectTypeCode, obj.SubjectTypeValue, m)
	sID := codeOrValueString(obj.SubjectIdCode, obj.SubjectIdValue, m)

	pType := codeOrValueString(obj.PredicateTypeCode, obj.PredicateTypeValue, m)
	pID := codeOrValueString(obj.PredicateIdCode, obj.PredicateIdValue, m)

	oType := codeOrValueString(obj.ObjectTypeCode, obj.ObjectTypeValue, m)
	oID := codeOrValueString(obj.ObjectIdCode, obj.ObjectIdValue, m)

	r.processRelation(sType, sID, pType, pID, oType, oID, obj)
}

func (r *runner) processRelation(sType, sID, pType, pID, oType, oID string, obj *RelationObject) {
	if sType == "" {
		r.logger.Warn("subject type is required", "relation.base", obj.Base, "id", sID)
		return
	}
	if sID == "" {
		r.logger.Warn("subject ID is required", "relation.base", obj.Base, "type", sType)
		return
	}

	if pType == "" {
		r.logger.Warn("predicate type is required", "relation.base", obj.Base, "id", pID)
		return
	}
	if pID == "" {
		r.logger.Warn("predicate ID is required", "relation.base", obj.Base, "type", pType)
		return
	}

	if oType == "" {
		r.logger.Warn("object type is required", "relation.base", obj.Base, "id", oID)
		return
	}
	if oID == "" {
		r.logger.Warn("object ID is required", "relation.base", obj.Base, "type", oType)
		return
	}

	r.manager.relations.AddRelation(
		models.NewRelation(
			models.NewEntity(sType, sID, r.manager.newAttributes()),
			models.NewEntity(pType, pID, r.manager.newAttributes()),
			models.NewEntity(oType, oID, r.manager.newAttributes()),
		),
	)
}
