package xsd

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// ToString converts a Golang variable into a string according to the XSD data-type.
//
// This function supports most of the common XSD data-types.
func ToString(data any, t string) (string, error) {
	switch t {
	case URIString, URINormalized, URIToken, URILanguage, URIAny, URISimple, URIAnyURI,
		PrefixString, PrefixNormalized, PrefixToken, PrefixLanguage, PrefixAny, PrefixSimple, PrefixAnyURI:
		return convert.AnyToString(data), nil
	case URIBoolean, PrefixBoolean:
		return convert.AnyToString(convert.AnyToBool(data)), nil
	case URIFloat, URIDouble, URIDecimal, PrefixFloat, PrefixDouble, PrefixDecimal:
		return convert.AnyToString(convert.AnyToFloat64(data)), nil
	case URIInteger, URILong, URINonPos, URINeg, URIInt, URIShort, URIByte,
		PrefixInteger, PrefixLong, PrefixNonPos, PrefixNeg, PrefixInt, PrefixShort, PrefixByte:
		return convert.AnyToString(convert.AnyToInt64(data)), nil
	case URINonNeg, URIPos, URIULong, URIUInt, URIUShort, URIUByte, URIDay, URIMonth, URIYear,
		PrefixNonNeg, PrefixPos, PrefixULong, PrefixUInt, PrefixYear, PrefixUShort, PrefixUByte, PrefixDay, PrefixMonth:
		return convert.AnyToString(convert.AnyToUint64(data)), nil
	case URIDuration, PrefixDuration:
		return toDurationString(data)
	case URIDateTime, PrefixDateTime:
		return convert.AnyToDateTime(data).Format(time.RFC3339Nano), nil
	case URITime, PrefixTime:
		return convert.AnyToDateTime(data).Format("15:04:05.999999999"), nil
	case URIDate, PrefixDate:
		return convert.AnyToDateTime(data).Format("2006-01-02"), nil
	case URIYearMonth, PrefixYearMonth:
		return convert.AnyToDateTime(data).Format("2006-01"), nil
	case URIMonthDay, PrefixMonthDay:
		return convert.AnyToDateTime(data).Format("01-02"), nil
	default:
		return "", fmt.Errorf("rdf: invalid data type: <%s>", t)
	}
}

func toDurationString(in any) (string, error) {
	var d time.Duration

	switch t := in.(type) {
	case time.Duration:
		d = t
	default:
		q, err := fromDurationString(convert.AnyToString(in))
		if err != nil {
			return "", err
		}
		d, _ = q.(time.Duration)
	}

	if d == 0 {
		return "PT0S", nil
	}

	var minus bool
	if d < 0 {
		d *= -1
		minus = true
	}

	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute
	d -= minutes * time.Minute
	seconds := float64(d) / float64(time.Second)

	out := bytes.Buffer{}

	if minus {
		out.WriteByte('-')
	}

	out.WriteString("PT")

	if hours > 0 {
		out.WriteString(strconv.FormatInt(int64(hours), 10))
		out.WriteByte('H')
	}

	if minutes > 0 {
		out.WriteString(strconv.FormatInt(int64(minutes), 10))
		out.WriteByte('M')
	}

	if seconds > 0 {
		out.WriteString(strconv.FormatFloat(seconds, 'f', -1, 64))
		out.WriteByte('S')
	}

	return out.String(), nil
}
