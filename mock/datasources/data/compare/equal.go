package compare

import (
	"bytes"
	"reflect"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// Equal returns true if the given data values are equal based on the given field type.
//
// For string type fields, the insensitive parameter can be used to perform the comparison in a case-insensitive manner.
// The default is to compare case-sensitive.
func Equal(t enums.FieldType, insensitive bool, v1, v2 any) bool {
	switch t {
	case enums.StringType, enums.URLType, enums.EmailType, enums.PhoneNrType, enums.IPAddressType:
		s1, s2 := convert.AnyToString(v1), convert.AnyToString(v2)
		if insensitive {
			return strings.EqualFold(s1, s2)
		}
		return s1 == s2

	case enums.IntegerType:
		return convert.AnyToInt64(v1) == convert.AnyToInt64(v2)
	case enums.UnsignedIntegerType:
		return convert.AnyToUint64(v1) == convert.AnyToUint64(v2)
	case enums.FloatType:
		return convert.AnyToFloat64(v1) == convert.AnyToFloat64(v2)
	case enums.BooleanType:
		return convert.AnyToBool(v1) == convert.AnyToBool(v2)

	case enums.DateType:
		d1 := convert.AnyToDateTime(v1)
		d2 := convert.AnyToDateTime(v2)
		year1, mon1, day1 := d1.Date()
		year2, mon2, day2 := d2.Date()
		return year1 == year2 && mon1 == mon2 && day1 == day2

	case enums.TimeType:
		d1 := convert.AnyToDateTime(v1)
		d2 := convert.AnyToDateTime(v2)
		d1 = time.Date(1970, 10, 10, d1.Hour(), d1.Minute(), d1.Second(), d1.Nanosecond(), d1.Location())
		d2 = time.Date(1970, 10, 10, d2.Hour(), d2.Minute(), d2.Second(), d2.Nanosecond(), d2.Location())
		return d1.Equal(d2)

	case enums.DateTimeType:
		return convert.AnyToDateTime(v1).Equal(convert.AnyToDateTime(v2))

	case enums.AnyType:
		return reflect.DeepEqual(v1, v2)

	case enums.ObjectType:
		b1, err1 := json.Marshal(v1)
		b2, err2 := json.Marshal(v2)
		return err1 == nil && err2 == nil && bytes.Equal(b1, b2)

	default:
		return false
	}
}
