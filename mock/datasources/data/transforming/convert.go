package transforming

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

func (r *runner) convert() any {
	v, _ := r.getValueAndType(1)
	if v == nil {
		return nil
	}
	return r.fixResultType(v)
}

func (r *runner) fixResultType(in any) any {
	switch r.transform.ResultType {
	case enums.IntegerType:
		return convert.AnyToInt64(in)
	case enums.UnsignedIntegerType:
		return convert.AnyToUint64(in)
	case enums.FloatType:
		return convert.AnyToFloat64(in)
	case enums.BooleanType:
		return convert.AnyToBool(in)
	case enums.DateType, enums.TimeType, enums.DateTimeType:
		return convert.AnyToDateTime(in)
	default:
		return convert.AnyToString(in)
	}
}
