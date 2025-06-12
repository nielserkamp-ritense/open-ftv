package transforming

import (
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
)

func (p *params) getValueAndType(i int) (any, enums.FieldType) {
	if v, tp, ok := p.getFieldValueAndType(i); ok {
		return v, tp
	}

	if v, tp, ok := p.getTransformValueAndType(i); ok {
		return v, tp
	}

	v := p.transform.InputValues[i]

	switch v.(type) {
	case string:
		return v, enums.StringType
	case bool:
		return v, enums.BooleanType
	case int, int8, int16, int32, int64:
		return v, enums.IntegerType
	case uint, uint8, uint16, uint32, uint64:
		return v, enums.UnsignedIntegerType
	case float32, float64:
		return v, enums.FloatType
	case time.Time, *time.Time:
		return v, enums.DateTimeType
	default:
		return v, enums.AnyType
	}
}

func (p *params) getFieldValueAndType(i int) (any, enums.FieldType, bool) {
	id, ok := p.transform.InputFields[i]
	if !ok {
		return nil, 0, false
	}

	f := p.transform.GetField(id)
	if f == nil {
		return nil, 0, false
	}

	var v any
	if p.qualified {
		v = p.rec[f.FQID()]
	} else {
		v = p.rec[f.ID]
	}
	return v, f.Type, true
}

func (p *params) getTransformValueAndType(i int) (any, enums.FieldType, bool) {
	id, ok := p.transform.InputTransforms[i]
	if !ok {
		return nil, 0, false
	}

	transform := p.transform.GetTransformation(id)
	if transform == nil {
		return nil, 0, false
	}

	var v any
	if p.qualified {
		v = p.rec[transform.FQID()]
	} else {
		v = p.rec[transform.ID]
	}

	if v != nil {
		return v, transform.ResultType, true
	}

	p2 := &params{rec: p.rec, qualified: p.qualified, transform: transform}
	return p2.run(), transform.ResultType, true
}
