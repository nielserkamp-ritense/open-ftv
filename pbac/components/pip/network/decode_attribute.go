package network

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/xsd"
)

func (r *runner) decodeAttribute(obj *AttributesMapping) {
	if data := findElement(splitKeys(obj.Base), r.data); data != nil {
		r.decodeAttributesData(data, obj, r.manager.attributes)
	}
}

func (r *runner) decodeAttributesData(data any, obj *AttributesMapping, set models.AttributeSet) {
	for k := range obj.Map {
		r.decodeAttributeData(obj.Base, data, obj.Map[k], set)
	}
}

func (r *runner) decodeAttributeData(base string, data any, obj *AttributeMapping, set models.AttributeSet) {
	if obj.ValueAsIs {
		r.processAttribute(base, obj.KeyValue, data, obj.TypeValue, set)
	} else {
		switch t := data.(type) {
		case []any:
			r.decodeAttributesSlice(base, t, obj, set)
		case map[string]any:
			r.decodeAttributeMap(base, t, obj, set)
		default:
			r.processAttribute(base, obj.KeyValue, data, obj.TypeValue, set)
		}
	}
}

func (r *runner) decodeAttributesSlice(base string, m []any, obj *AttributeMapping, set models.AttributeSet) {
	for i := range m {
		r.decodeAttributeData(base, m[i], obj, set)
	}
}

func (r *runner) decodeAttributeMap(base string, m map[string]any, obj *AttributeMapping, set models.AttributeSet) {
	key := codeOrValueString(obj.KeyField, obj.KeyValue, m)
	value := m[obj.ValueField]
	tp := codeOrValueString(obj.TypeField, obj.TypeValue, m)
	r.processAttribute(base, key, value, tp, set)
}

func (r *runner) processAttribute(base, key string, value any, tp string, set models.AttributeSet) {
	if key == "" {
		r.logger.Warn("attribute key is required", "attribute.base", base, "value", value, "type", tp)
		return
	}
	if value == nil {
		r.logger.Warn("attribute value is required", "attribute.base", base, "key", key, "type", tp)
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
