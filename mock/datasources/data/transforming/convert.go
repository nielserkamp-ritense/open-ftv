package transforming

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

func (p *runner) convert() any {
	v, _ := p.getValueAndType(1)

	switch p.transform.ResultType {
	case enums.IntegerType:
		return convert.AnyToInt64(v)
	case enums.UnsignedIntegerType:
		return convert.AnyToUint64(v)
	case enums.FloatType:
		return convert.AnyToFloat64(v)
	case enums.BooleanType:
		return convert.AnyToBool(v)
	case enums.DateType, enums.TimeType, enums.DateTimeType:
		return convert.AnyToDateTime(v)
	default:
		return convert.AnyToString(v)
	}
}
