package transforming

import (
	"strings"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
)

func (r *runner) getValueAndType(i int) (any, enums.FieldType) {
	if v, tp, ok := r.getFieldValueAndType(i); ok {
		return v, tp
	}

	if v, tp, ok := r.getTransformValueAndType(i); ok {
		return v, tp
	}

	return r.testType(r.transform.InputValues[i])
}

func (r *runner) getFieldValueAndType(i int) (any, enums.FieldType, bool) {
	id, ok := r.transform.InputFields[i]
	if !ok {
		return nil, 0, false
	}

	f := r.transform.GetField(id)
	if f == nil {
		return nil, 0, false
	}

	var v any
	if r.qualified {
		v = r.rec[f.FQID()]
	} else {
		v = r.rec[f.ID]
	}
	return v, f.Type, true
}

func (r *runner) getTransformValueAndType(i int) (any, enums.FieldType, bool) {
	id, ok := r.transform.InputTransforms[i]
	if !ok {
		return nil, 0, false
	}

	transform := r.transform.GetTransformation(id)
	if transform == nil {
		return nil, 0, false
	}

	var v any
	if r.qualified {
		v = r.rec[transform.FQID()]
	}

	if !r.qualified || v == nil {
		v = r.rec[transform.ID]
	}

	if v != nil {
		return v, transform.ResultType, true
	}

	// if the transformation hasn't been executed yet, we'll force it here.
	p2 := &runner{rec: r.rec, qualified: r.qualified, transform: transform}
	return p2.run(), transform.ResultType, true
}

func (r *runner) testType(in any) (any, enums.FieldType) {
	switch t := in.(type) {
	case string:
		return r.testParameter(t)
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

func (r *runner) testParameter(s string) (any, enums.FieldType) {
	// a string like ":xyz:" represents a parameter with the key "xyz".
	if strings.HasPrefix(s, ":") && strings.HasSuffix(s, ":") {
		if v, ok := r.params[s[1:len(s)-1]]; ok {
			return r.testType(v)
		}
		return nil, enums.AnyType
	}
	return s, enums.StringType
}
