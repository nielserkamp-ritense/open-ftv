package network

import (
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/xsd"
)

func (r *runner) decodeAttribute(obj *AttributesMapping) {
	if data := findElement(splitKeys(obj.Base), r.data); data != nil {
		r.decodeAttributesData(data, obj, r.manager.addAttribute)
	}
}

func (r *runner) decodeAttributesData(data any, obj *AttributesMapping, add models.AddAttribute) {
	for k := range obj.Map {
		r.decodeAttributeData(obj.Base, data, obj.Map[k], add)
	}
}

func (r *runner) decodeAttributeData(base string, data any, obj *AttributeMapping, add models.AddAttribute) {
	if obj.ValueAsIs {
		r.processAttribute(base, r.addParents(obj.KeyValue, obj.KeyParents), data, obj.TypeValue, add)
	} else {
		switch t := data.(type) {
		case []any:
			r.decodeAttributesSlice(base, t, obj, add)
		case map[string]any:
			r.decodeAttributeMap(base, t, obj, add)
		default:
			r.processAttribute(base, r.addParents(obj.KeyValue, obj.KeyParents), data, obj.TypeValue, add)
		}
	}
}

func (r *runner) decodeAttributesSlice(base string, m []any, obj *AttributeMapping, add models.AddAttribute) {
	for i := range m {
		r.decodeAttributeData(base, m[i], obj, add)
	}
}

func (r *runner) decodeAttributeMap(base string, m map[string]any, obj *AttributeMapping, add models.AddAttribute) {
	key := codeOrValueString(obj.KeyField, obj.KeyValue, m)
	value := codeOrValue(obj.ValueField, obj.ValueValue, m)
	tp := codeOrValueString(obj.TypeField, obj.TypeValue, m)
	r.processAttribute(base, r.addParents(key, obj.KeyParents), value, tp, add)
}

func (r *runner) addParents(key string, parents []string) string {
	return strings.Join(append(parents, key), ".")
}

func (r *runner) processAttribute(base, key string, value any, tp string, add models.AddAttribute) {
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

	add(key, v1, value, tp)
}
