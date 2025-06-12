package compare

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// Greater returns true if the first given data value is lexically or numerically greater.
// The type of comparison is based on the given field type.
//
// For string type fields, the insensitive parameter can be used to perform the comparison in a case-insensitive manner.
// The default is to compare case-sensitive.
func Greater(t enums.FieldType, insensitive bool, v1, v2 any) bool {
	switch t {
	case enums.IntegerType:
		return convert.AnyToInt64(v1) > convert.AnyToInt64(v2)
	case enums.UnsignedIntegerType:
		return convert.AnyToUint64(v1) > convert.AnyToUint64(v2)
	case enums.FloatType:
		return convert.AnyToFloat64(v1) > convert.AnyToFloat64(v2)
	case enums.BooleanType:
		return convert.AnyToBool(v1) && !convert.AnyToBool(v2)

	case enums.DateType, enums.TimeType, enums.DateTimeType:
		return convert.AnyToDateTime(v1).After(convert.AnyToDateTime(v2))

	case enums.AnyType:
		return fmt.Sprintf("%v", v1) > fmt.Sprintf("%v", v2)

	case enums.ObjectType:
		b1, err1 := json.Marshal(v1)
		b2, err2 := json.Marshal(v2)
		return err1 == nil && (err2 != nil || bytes.Compare(b1, b2) > 0)

	default:
		s1, s2 := convert.AnyToString(v1), convert.AnyToString(v2)
		if insensitive {
			return strings.ToLower(s1) > strings.ToLower(s2)
		}
		return s1 > s2
	}
}
