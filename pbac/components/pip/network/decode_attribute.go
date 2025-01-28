package network

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/xsd"
)

func (r *runner) decodeAttribute(obj *AttributeObject) {
	if data := findElement(splitKeys(obj.Base), r.data); data != nil {
		r.decodeAttributeData(data, obj, r.manager.attributes)
	}
}

func (r *runner) decodeAttributeData(data any, obj *AttributeObject, set models.AttributeSet) {
	if obj.ValueAsIs {
		r.processAttribute(obj.KeyValue, data, obj.TypeValue, obj, set)
	} else {
		switch t := data.(type) {
		case []any:
			r.decodeAttributesSlice(t, obj, set)
		case map[string]any:
			r.decodeAttributeMap(t, obj, set)
		default:
			r.processAttribute(obj.KeyValue, data, obj.TypeValue, obj, set)
		}
	}
}

func (r *runner) decodeAttributesSlice(m []any, obj *AttributeObject, set models.AttributeSet) {
	for i := range m {
		r.decodeAttributeData(m[i], obj, set)
	}
}

func (r *runner) decodeAttributeMap(m map[string]any, obj *AttributeObject, set models.AttributeSet) {
	key := codeOrValueString(obj.KeyCode, obj.KeyValue, m)
	value := m[obj.ValueCode]
	tp := codeOrValueString(obj.TypeCode, obj.TypeValue, m)
	r.processAttribute(key, value, tp, obj, set)
}

func (r *runner) processAttribute(key string, value any, tp string, obj *AttributeObject, set models.AttributeSet) {
	if key == "" {
		r.logger.Warn("attribute key is required", "attribute.base", obj.Base, "value", value, "type", tp)
		return
	}
	if value == nil {
		r.logger.Warn("attribute value is required", "attribute.base", obj.Base, "key", key, "type", tp)
		return
	}

	v1 := value
	if tp != "" {
		if v2, err := xsd.FromString(convert.AnyToString(value), tp); err == nil {
			v1 = v2
		}
	}

	set.AddOriginalAttribute(key, v1, value, tp)
}
