package network

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

func (r *runner) decodeEntity(obj *EntityObject) {
	if data := findElement(splitKeys(obj.Base), r.data); data != nil {
		r.decodeEntityData(data, obj)
	}
}

func (r *runner) decodeEntityData(data any, obj *EntityObject) {
	if obj.IdFromValue {
		r.processEntity(obj.TypeValue, convert.AnyToString(data), nil, nil, obj)
	} else {
		switch t := data.(type) {
		case []any:
			r.decodeEntitiesSlice(t, obj)
		case map[string]any:
			r.decodeEntityMap(t, obj)
		default:
			r.processEntity(obj.TypeValue, convert.AnyToString(data), nil, nil, obj)
		}
	}
}

func (r *runner) decodeEntitiesSlice(m []any, obj *EntityObject) {
	for i := range m {
		r.decodeEntityData(m[i], obj)
	}
}

func (r *runner) decodeEntityMap(m map[string]any, obj *EntityObject) {
	tp := codeOrValueString(obj.TypeCode, obj.TypeValue, m)
	id := codeOrValueString(obj.IdCode, obj.IdValue, m)

	// determine optional attributes.
	attrs := r.manager.newAttributes()
	for _, attrObj := range obj.Attributes {
		if data := findElement(splitKeys(attrObj.Base), m); data != nil {
			r.decodeAttributeData(data, attrObj, attrs)
		}
	}

	// determine optional parents.
	var parents []string
	if obj.ParentsCode != "" {
		switch t := m[obj.ParentsCode].(type) {
		case []string:
			parents = t
		case []any:
			for i := range t {
				parents = append(parents, convert.AnyToString(t[i]))
			}
		default:
			parents = []string{convert.AnyToString(t)}
		}
	}

	r.processEntity(tp, id, attrs, parents, obj)
}

func (r *runner) processEntity(tp string, id string, attrs models.AttributeSet, parents []string, obj *EntityObject) {
	if tp == "" {
		r.logger.Warn("entity type is required", "entity.base", obj.Base, "id", id)
		return
	}
	if id == "" {
		r.logger.Warn("entity ID is required", "entity.base", obj.Base, "type", tp)
		return
	}

	if attrs == nil {
		attrs = r.manager.newAttributes()
	}

	r.manager.entities.AddEntity(models.NewEntity(tp, id, attrs, parents...))
}
