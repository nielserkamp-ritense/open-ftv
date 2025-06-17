package transforming

import (
	"strings"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
)

func (p *runner) getValueAndType(i int) (any, enums.FieldType) {
	if v, tp, ok := p.getFieldValueAndType(i); ok {
		return v, tp
	}

	if v, tp, ok := p.getTransformValueAndType(i); ok {
		return v, tp
	}

	return p.testType(p.transform.InputValues[i])
}

func (p *runner) getFieldValueAndType(i int) (any, enums.FieldType, bool) {
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

func (p *runner) getTransformValueAndType(i int) (any, enums.FieldType, bool) {
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

	p2 := &runner{rec: p.rec, qualified: p.qualified, transform: transform}
	return p2.run(), transform.ResultType, true
}

func (p *runner) testType(in any) (any, enums.FieldType) {
	switch t := in.(type) {
	case string:
		return p.testParameter(t)
	case bool:
		return t, enums.BooleanType
	case int, int8, int16, int32, int64:
		return t, enums.IntegerType
	case uint, uint8, uint16, uint32, uint64:
		return t, enums.UnsignedIntegerType
	case float32, float64:
		return t, enums.FloatType
	case time.Time, *time.Time:
		return t, enums.DateTimeType
	default:
		return t, enums.AnyType
	}
}

func (p *runner) testParameter(s string) (any, enums.FieldType) {
	// a string like ":xyz:" represents a parameter with the key "xyz".
	if strings.HasPrefix(s, ":") && strings.HasSuffix(s, ":") {
		if v, ok := p.params[s[1:len(s)-1]]; ok {
			return p.testType(v)
		}
		return nil, enums.AnyType
	}
	return s, enums.StringType
}
