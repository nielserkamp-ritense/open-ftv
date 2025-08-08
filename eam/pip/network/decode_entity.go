package network

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

func (r *runner) decodeEntity(obj *EntityMapping) {
	if data := findElement(splitKeys(obj.Base), r.data); data != nil {
		r.decodeEntityData(data, obj)
	}
}

func (r *runner) decodeEntityData(data any, obj *EntityMapping) {
	if obj.IDFromValue {
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

func (r *runner) decodeEntitiesSlice(m []any, obj *EntityMapping) {
	for i := range m {
		r.decodeEntityData(m[i], obj)
	}
}

func (r *runner) decodeEntityMap(m map[string]any, obj *EntityMapping) {
	tp := codeOrValueString(obj.TypeField, obj.TypeValue, m)
	id := codeOrValueString(obj.IDField, obj.IDValue, m)

	// determine optional attributes.
	attrs := models.NewAttributeSet()
	for _, attrObj := range obj.Attributes {
		if data := findElement(splitKeys(attrObj.Base), m); data != nil {
			r.decodeAttributesData(data, attrObj, attrs.AddAttribute)
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

func (r *runner) processEntity(tp string, id string, attrs *models.AttributeSet, parents []string, obj *EntityMapping) {
	if tp == "" {
		r.logger.Warn("entity type is required", "entity.base", obj.Base, "id", id)
		return
	}
	if id == "" {
		r.logger.Warn("entity ID is required", "entity.base", obj.Base, "type", tp)
		return
	}

	if attrs == nil {
		attrs = models.NewAttributeSet()
	}

	r.manager.addEntity(models.NewEntity(tp, id, attrs, parents...))
}
